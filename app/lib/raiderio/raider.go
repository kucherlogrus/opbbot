package raiderio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const api_url = "https://raider.io/api/v1/"

type RaiderApi struct {
	api_url string
}

func CreateApi() *RaiderApi {
	return &RaiderApi{
		api_url: api_url,
	}
}

func (ra *RaiderApi) GetCurrentAffixes() (*AutoGeneratedActiveAffixes, error) {
	params := map[string]interface{}{"locale": "ru"}
	raw_resp, err := ra.__callApi("GET", "mythic-plus/affixes", params)
	if err != nil {
		return nil, err
	}

	var result AutoGeneratedActiveAffixes
	if err = json.Unmarshal(raw_resp, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
		return nil, err
	}

	return &result, nil
}

func (ra *RaiderApi) GetUserInfo(realm string, name string) (*AutoGeneratedPlayerInfo, error) {
	params := map[string]interface{}{
		"realm":  realm,
		"name":   name,
		"fields": "mythic_plus_scores_by_season:current,mythic_plus_ranks,gear,mythic_plus_best_runs",
	}
	raw_resp, err := ra.__callApi("GET", "characters/profile", params)
	if err != nil {
		return nil, err
	}

	var result AutoGeneratedPlayerInfo
	if err = json.Unmarshal(raw_resp, &result); err != nil {
		return nil, fmt.Errorf("Can not unmarshal JSON", err)
	}

	return &result, err
}

func (ra *RaiderApi) __callApi(method string, endpoint string, params map[string]interface{}) (raw_resp []byte, err error) {
	var resp *http.Response
	if method == "GET" {
		qparams := url.Values{}
		if params != nil {
			for key, value := range params {
				qparams.Add(key, fmt.Sprintf("%v", value))
			}
		}

		qparams.Add("region", "eu")
		full_url := ra.api_url + endpoint + "?" + qparams.Encode()
		resp, err = http.Get(full_url)
		defer resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if resp.StatusCode == 400 {
			var error_message AutoGeneratedErrorMessage
			if err = json.Unmarshal(body, &error_message); err != nil {
				return nil, fmt.Errorf("Can not unmarshal JSON", err)
			}
			return nil, fmt.Errorf(error_message.Message)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("Can't handle request, code %s", resp.StatusCode)
		}

		return body, err

	}
	return nil, fmt.Errorf("Method : %s doesn't support yet", method)
}
