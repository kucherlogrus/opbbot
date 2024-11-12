package battlenet

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"opb_bot/lib/db"
	"opb_bot/lib/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//curl -u client:secret -d grant_type=client_credentials https://oauth.battle.net/token

const oauth_url = "https://oauth.battle.net/token"

const debug = false

const api_url = "https://eu.api.blizzard.com"
const news_url = "https://worldofwarcraft.com/en-gb/news"

type Affix struct {
	ID          int
	Name        string
	EngName     string
	Description string
}

type Dungeon struct {
	ID          int
	Name        string
	EngName     string
	Description string
}

type Battlenet struct {
	Affixes_map map[string]Affix
	Dungeon_map map[string]Dungeon
	Token       *db.ServiceDB
	db_instance *db.DBHandler
}

type BattleNetNews struct {
	ID      int
	Tittle  string
	URL     string
	Timestr string
}

func (bn *Battlenet) InitBattlenetApi(db_instance *db.DBHandler) error {
	bn.db_instance = db_instance
	service, err := db_instance.GetService("battlenet")
	if err != nil {
		return fmt.Errorf("Can't init battle net api: ", err)
	}
	bn.Token = service
	bn.Affixes_map = map[string]Affix{}
	bn.loadAffixes()
	if len(bn.Affixes_map) == 0 {
		var affixes []Affix
		affixes, err = bn.getAffixesFromBattleNet()
		if err != nil {
			return fmt.Errorf("Can't init battle net api: ", err)
		}
		err = bn.saveAffixes(affixes)
		for _, affix := range affixes {
			bn.Affixes_map[affix.EngName] = affix
		}
		if err != nil {
			return fmt.Errorf("Can't init battle net api: ", err)
		}
	}
	bn.Dungeon_map = map[string]Dungeon{}
	bn.loadDungeons()
	if len(bn.Dungeon_map) == 0 {
		var dungeons []Dungeon
		dungeons, err = bn.getDungeonsFromBattleNet()
		if err != nil {
			return fmt.Errorf("Can't init battle net api: ", err)
		}
		err = bn.saveDungeons(dungeons)
		for _, dungeon := range dungeons {
			bn.Dungeon_map[dungeon.EngName] = dungeon
		}
		if err != nil {
			return fmt.Errorf("Can't init battle net api: ", err)
		}
	}
	return nil
}

func (bn *Battlenet) ReloadBattleNetData() error {
	bn.refreshToken()
	var dungeons []Dungeon
	dungeons, err := bn.getDungeonsFromBattleNet()
	if err != nil {
		return fmt.Errorf("Can't init battle net api: ", err)
	}
	err = bn.saveDungeons(dungeons)
	for _, dungeon := range dungeons {
		bn.Dungeon_map[dungeon.EngName] = dungeon
	}
	if err != nil {
		return fmt.Errorf("Can't init battle net api: ", err)
	}
	return nil
}

func (bn *Battlenet) TokenVerify() (err error) {
	err = bn.checkTokenValid()
	return
}

func (bn *Battlenet) refreshToken() (err error) {
	resp, err := GetBattleNetToken(bn.Token.Client, bn.Token.Secret)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	expires := time.Now().Add(time.Duration(resp.ExpiresIn))
	fmt.Println("Battle.net: refresh token", resp.AccessToken, expires)
	bn.db_instance.RefreshAccessToken("battlenet", resp.AccessToken, expires)
	service, err := bn.db_instance.GetService("battlenet")
	if err != nil {
		return fmt.Errorf("Can't init battle net api: ", err)
	}
	bn.Token = service
	return nil
}

func (bn *Battlenet) loadAffixes() {
	row, err := bn.db_instance.GetBattleNetAffixesRow()
	if err != nil {
		return
	}
	defer row.Close()
	for row.Next() {
		var affix Affix
		row.Scan(&affix.ID, &affix.Name, &affix.EngName, &affix.Description)
		bn.Affixes_map[affix.EngName] = affix
	}
}

func (bn *Battlenet) saveAffixes(affixes []Affix) error {
	fmt.Println("Battle.net: save affixes to database")
	_, err := bn.db_instance.Connection.Exec("DELETE FROM wowaffixes WHERE id not null")
	if err != nil {
		fmt.Println(err)
		return err
	}
	var statement *sql.Stmt
	statement, err = bn.db_instance.Connection.Prepare("INSERT into wowaffixes (id, name, engname, description) VALUES (?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, affix := range affixes {
		if affix.ID <= 0 {
			continue
		}
		fmt.Print("Save affix: ")
		utils.PrintType(affix)
		_, err = statement.Exec(affix.ID, affix.Name, affix.EngName, affix.Description)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil

}

func (bn *Battlenet) loadDungeons() {
	row, err := bn.db_instance.GetBattleNetDungeonRow()
	if err != nil {
		return
	}
	defer row.Close()
	for row.Next() {
		var dungeon Dungeon
		row.Scan(&dungeon.ID, &dungeon.Name, &dungeon.EngName, &dungeon.Description)
		bn.Dungeon_map[dungeon.EngName] = dungeon
	}
}

func (bn *Battlenet) saveDungeons(dungeons []Dungeon) error {
	fmt.Println("Battle.net: save dungeons to database")
	_, err := bn.db_instance.Connection.Exec("DELETE FROM dungeon WHERE id not null")
	if err != nil {
		fmt.Println(err)
		return err
	}
	var statement *sql.Stmt
	statement, err = bn.db_instance.Connection.Prepare("INSERT into dungeon (id, name, engname, description) VALUES (?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, dungeon := range dungeons {
		if dungeon.ID <= 0 {
			continue
		}
		_, err = statement.Exec(dungeon.ID, dungeon.Name, dungeon.EngName, dungeon.Description)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil

}

func (bn *Battlenet) getAffixesFromBattleNet() ([]Affix, error) {
	fmt.Println("Battle.net: get affixes from server.")
	err := bn.checkTokenValid()
	if err != nil {
		return nil, err
	}
	resp, err := bn.__callApi("GET", "/data/wow/keystone-affix/index", nil, bn.Token.Access_token, true)
	var result AutoGeneratedAffixes
	if err = json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("Can not unmarshal JSON", err)
	}
	affixes := make([]Affix, len(result.Affixes))
	for _, affix_el := range result.Affixes {
		affix := Affix{}
		affix.ID = affix_el.ID
		affix.Name = affix_el.Name.RuRU
		affix.EngName = affix_el.Name.EnGB
		affix_def_battle_net, _ := bn.getAffixDefinitionFromBattleNet(affix_el.ID)
		affix.Description = affix_def_battle_net.Description
		affixes = append(affixes, affix)
	}
	return affixes, nil
}

func (bn *Battlenet) getAffixDefinitionFromBattleNet(id int) (*AutoGeneratedAfixDef, error) {
	var afixdef AutoGeneratedAfixDef
	err := bn.checkTokenValid()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	resp, err := bn.__callApi("GET", fmt.Sprintf("/data/wow/keystone-affix/%d", id), nil, bn.Token.Access_token, false)
	if err = json.Unmarshal(resp, &afixdef); err != nil {
		return nil, fmt.Errorf("Can not unmarshal JSON: ", err)
	}
	return &afixdef, err
}

func (bn *Battlenet) getDungeonsFromBattleNet() ([]Dungeon, error) {
	fmt.Println("Battle.net: get dungeons from server.")
	err := bn.checkTokenValid()
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"namespace": "dynamic-eu",
	}

	resp, err := bn.__callApi("GET", "/data/wow/mythic-keystone/dungeon/index", params, bn.Token.Access_token, true)
	var result AutoGeneratedDungeons
	if err = json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("Can not unmarshal JSON", err)
	}
	dungeons := make([]Dungeon, len(result.Dungeons))
	for _, dungeon_el := range result.Dungeons {
		dungeon := Dungeon{}
		dungeon.ID = dungeon_el.ID
		dungeon.Name = dungeon_el.Name.RuRU
		dungeon.EngName = dungeon_el.Name.EnGB
		dungeon.Description = ""
		dungeons = append(dungeons, dungeon)
	}
	return dungeons, nil
}

func GetBattleNetToken(clientID string, secretID string) (resp *AutoGeneratedOauthResp, err error) {
	client := &http.Client{Timeout: time.Second * 10}
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", oauth_url, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.SetBasicAuth(clientID, secretID)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()
	var result AutoGeneratedOauthResp
	body, err := ioutil.ReadAll(response.Body)
	resp_string := string(body)
	if strings.Contains(resp_string, "error") {
		var err_oauth AutoGeneratedOauthErr
		if err = json.Unmarshal(body, &err_oauth); err != nil {
			return nil, fmt.Errorf(err_oauth.ErrorDescription)
		}
	} else {
		if err = json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("Can not unmarshal JSON", err)
		}
	}

	return &result, err
}

func (bn *Battlenet) __callApi(method string, endpoint string, params map[string]interface{}, token string, all_locale bool) (raw_resp []byte, err error) {
	var resp *http.Response
	if method == "GET" {
		qparams := url.Values{}
		if params != nil {
			for key, value := range params {
				qparams.Add(key, fmt.Sprintf("%v", value))
			}
		}
		qparams.Add("region", "eu")
		qparams.Add("access_token", token)
		if !all_locale {
			qparams.Add("locale", "ru_RU")
		}
		if !qparams.Has("namespace") {
			qparams.Add("namespace", "static-eu")
		}
		full_url := api_url + endpoint + "?" + qparams.Encode()
		request, err_resp := http.NewRequest("GET", full_url, nil)
		if err_resp != nil {
			return nil, err_resp
		}
		request.Header.Add("Authorization", "Bearer "+token)
		resp, err = http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			var error_message AutoGeneratedBattlenetErrorMessage
			if err = json.Unmarshal(body, &error_message); err != nil {
				return nil, fmt.Errorf("Can not unmarshal JSON", err)
			}
			return nil, fmt.Errorf(error_message.Detail)
		}
		return body, err

	}
	return nil, fmt.Errorf("Method : %s doesn't support yet", method)
}

func (bn *Battlenet) checkTokenValid() error {
	now := time.Now()
	expire_at := *bn.Token.Expire_at
	if now.After(expire_at) || now.Sub(expire_at) >= 300 {
		err := bn.refreshToken()
		if err != nil {
			return err
		}
	}
	return nil
}

type newTime struct {
	Iso      string `json:"iso8601"`
	Relative bool   `json:"relative"`
}

func (bn *Battlenet) GetLastNews(last_time string) ([]BattleNetNews, error) {
	resp, err := http.Get(news_url)

	if err != nil {
		return nil, fmt.Errorf("Can't get news from battle net. ", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Can't get news from battle net. ", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error: %s ", err)
	}

	regex, _ := regexp.Compile("news\\/(\\d+)\\/")

	value_t, err := time.Parse(time.RFC3339, last_time)
	var news []BattleNetNews

	var accept = true

	doc.Find(".NewsBlog").Each(func(i int, s *goquery.Selection) {
		if !accept {
			return
		}

		var id int
		var tittle string
		var url_link string
		var new_time string

		time_base_el := s.Find(".NewsBlog-date.LocalizedDateMount")
		if time_base_el != nil {
			props, exist := time_base_el.Attr("data-props")
			if exist {
				var nt newTime
				if _err := json.Unmarshal([]byte(props), &nt); _err != nil {
					fmt.Println("Can not unmarshal JSON: ", _err)
				}
				new_time = nt.Iso
				new_time_t, err := time.Parse(time.RFC3339, new_time)
				if err != nil {
					return
				}

				if new_time_t.Before(value_t) {
					accept = false
					return
				}
			} else {
				return
			}
		}

		title_el := s.Find(".NewsBlog-title")
		if title_el != nil {
			tittle = title_el.Text()
		} else {
			return
		}

		link_el := s.Find(".NewsBlog-link")

		if link_el != nil {
			link, exist := link_el.Attr("href")
			if exist {
				match := regex.FindStringSubmatch(link)
				if len(match) == 2 {
					id, err = strconv.Atoi(match[1])
					if err != nil {
						return
					}
				}
				url_link = "https://worldofwarcraft.com" + link
			} else {
				return
			}
		}
		news = append(news, BattleNetNews{id, tittle, url_link, new_time})

	})

	return news, nil

}

func (bn *Battlenet) GetNewFromUrl(url_link string) (text string, error error) {
	resp, err := http.Get(url_link)

	if err != nil {
		return "", fmt.Errorf("Can't get new from battle net. ", err)
	}

	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Can't get news from battle net. %d\n", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error: %s ", err)
	}
	var in_last_news = false
	var can_append = true
	text = ""
	hr_count := 0
	if strings.Contains(url_link, "hotfixes") {
		doc.Find(".detail").Children().Each(func(i int, s *goquery.Selection) {
			if !can_append {
				return
			}
			name := goquery.NodeName(s)
			if name == "h2" && !in_last_news {
				text += bn.parse_text_element(s)
				in_last_news = true
				return
			}

			if name == "hr" && in_last_news {
				hr_count += 1
				if hr_count == 2 {
					in_last_news = false
					can_append = false
					return
				}
			}

			if in_last_news {
				text += bn.parse_text_element(s)
			}

		})
		return
	}

	if strings.Contains(url_link, "потасовке-на-этой") {
		return
	}

	doc.Find("#blog").Children().Each(func(i int, s *goquery.Selection) {
		text += bn.parse_text_element(s)
	})

	return
}

func (bn *Battlenet) parse_text_element(s *goquery.Selection) string {

	children_count := s.Children().Length()
	name := goquery.NodeName(s)
	if children_count == 0 {
		tag_text := bn.tag_handle(s)
		if debug {
			fmt.Printf("<%s>: %q\n", name, tag_text)
		}
		return tag_text
	}

	var text = ""
	if children_count > 0 {
		s.Contents().Each(func(i int, s *goquery.Selection) {
			text += bn.parse_text_element(s)
		})
		return text
	}
	return text

}

func (bn *Battlenet) tag_handle(s *goquery.Selection) string {
	var text = ""
	name := goquery.NodeName(s)

	raw_text := s.Text()
	if len(raw_text) < 4 {
		return text
	}

	switch name {
	case "strong":
		text += "**__" + raw_text + "__**\n"
	case "li":
		text += "  * " + raw_text + "\n"
	case "em":
		text += "  * " + raw_text + "\n"
	case "a":
		ref, exist := s.Attr("href")
		if exist {
			ref = strings.Replace(ref, "https://urldefense.com/v3/__", "", -1)
			text += raw_text + ": " + ref + "\n"
		} else {
			text += "  * " + raw_text + "\n"
		}
	case "#text":
		text += raw_text + "\n"
	default:
		text += s.Text() + "\n"

	}
	if goquery.NodeName(s.Parent()) == "ul" {
		text = "  " + text
	}
	text = bn._textNormalize(text)
	return text
}

func (bn *Battlenet) _textNormalize(text string) string {
	text = strings.Replace(text, "\t", "", -1)
	text = strings.Replace(text, "\n\n", "", -1)
	var chars = []string{"[", "]", ","}
	for _, c := range chars {
		if strings.HasPrefix(text, c) {
			text = strings.Replace(text, c, "", 1)
		}
	}
	if text == "\n" {
		text = ""
	}
	return text
}
