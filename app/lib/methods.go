package lib

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"opb_bot/lib/egs"
	"regexp"
	"strings"
	"time"
)

func (handler *BotHandler) Help(s *discordgo.Session, m *discordgo.MessageCreate) {
	text_message := handler._getHelpMessage()
	s.ChannelMessageSend(m.ChannelID, text_message)
}

func (handler *BotHandler) Clear(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID != test_channel_id {
		return
	}
	messages_raw, _ := s.ChannelMessages(m.ChannelID, 0, m.ID, "", "")
	for _, msg := range messages_raw {
		s.ChannelMessageDelete(m.ChannelID, msg.ID)
	}
}

func (handler *BotHandler) Egsupdates(s *discordgo.Session, m *discordgo.MessageCreate) {
	current_free_games, err := egs.ParseFreeEgsGamesUrls()
	if err != nil {
		fmt.Println("Error ParseFreeEgsGamesUrls,", err)
	}
	chat_free_games := map[string]string{}
	r, _ := regexp.Compile("#id:(.*)\n")
	messages_raw, _ := s.ChannelMessages(m.ChannelID, 0, m.ID, "", "")
	for _, msg := range messages_raw {
		match := r.FindStringSubmatch(msg.Content)
		if len(match) == 2 {
			chat_free_games[match[1]] = msg.ID
		}
	}
	//remove old games
	for key, value := range chat_free_games {
		_, exists := current_free_games[key]
		if !exists {
			s.ChannelMessageDelete(m.ChannelID, value)
		}
	}
	//add new games
	for key, game := range current_free_games {
		_, exists := chat_free_games[key]
		if !exists {
			game_string := fmt.Sprintf("\n#id:%s\nНазвание игры: %s\nОписание: %s\nURL: %s\n", game.ID, game.Title, game.Description, game.URL)
			s.ChannelMessageSend(m.ChannelID, game_string)
		}
	}
}

func (handler *BotHandler) Wowupdate(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := handler.raider.GetCurrentAffixes()
	if err != nil {
		fmt.Println("Error raiderio api call, ", err)
	}
	affixes := resp.AffixDetails
	var text_affixes []string
	text_affixes = append(text_affixes, "Аффиксы:\n-----------")

	for _, aff := range affixes {
		text_affixes = append(text_affixes, fmt.Sprintf("%s : %s\n", aff.Name, aff.Description))
	}
	message := strings.Join(text_affixes, "\n")
	s.ChannelMessageSend(m.ChannelID, message)
}

func (handler *BotHandler) Raider(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.Content)
	params := strings.Split(m.Content, " ")
	var realm, name string
	params_count := len(params)
	if params_count == 1 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Ошибка: Необходимо указать имя персонажа или сервер + имя\n"))
		return
	}
	if params_count == 2 {
		name = params[1]
		realm = "gordunni"
	} else {
		name = params[2]
		realm = params[1]
	}
	fmt.Println(params)
	result, err := handler.raider.GetUserInfo(realm, name)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Ошибка: %s", err))
	}
	var message = ""
	message += fmt.Sprintf("Имя: %s\n", result.Name)
	message += fmt.Sprintf("Сервер: %s\n", result.Realm)
	message += fmt.Sprintf("Рейтинг: %d\n", int(result.MythicPlusScoresBySeason[0].Scores.All))
	r := result.LastCrawledAt
	scan_time := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", r.Day(), r.Month(), r.Year(), r.Hour(), r.Minute(), r.Second())
	message += fmt.Sprintf("Дата обновления raider.io: %s\n", scan_time)
	message += fmt.Sprintf("-----------------------------------\n")
	for _, instance := range result.MythicPlusBestRuns {
		t := instance.CompletedAt
		complete_at := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second())
		tm := instance.ClearTimeMs
		sec := tm / 1000
		msec := tm % 1000
		tm_time := time.Unix(int64(sec), int64(msec*int(time.Millisecond)))
		tm_time_str := fmt.Sprintf("%d минут %d секунд", tm_time.Minute(), tm_time.Second())
		dungeon_name := handler.battlenet.Dungeon_map[instance.Dungeon]
		message += fmt.Sprintf("%s **__%d__**, пройден %s за %s. Аффиксы: ", dungeon_name.Name, instance.MythicLevel, complete_at, tm_time_str)
		for index, afix := range instance.Affixes {
			afix_name := handler.battlenet.Affixes_map[afix.Name].Name
			if index == len(instance.Affixes)-1 {

				message += fmt.Sprintf("%s\n", afix_name)
				continue
			}
			message += fmt.Sprintf("%s, ", afix_name)
		}
	}

	s.ChannelMessageSend(m.ChannelID, message)
}

func (handler *BotHandler) Affixes(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := handler.raider.GetCurrentAffixes()
	if err != nil {
		fmt.Println("Error raiderio api call, ", err)
	}
	affixes := resp.AffixDetails
	var text_affixes []string
	text_affixes = append(text_affixes, "Аффиксы:\n-----------")

	for _, aff := range affixes {
		text_affixes = append(text_affixes, fmt.Sprintf("%s : %s\n", aff.Name, aff.Description))
	}
	message := strings.Join(text_affixes, "\n")

	s.ChannelMessageSend(m.ChannelID, message)
}

func (handler *BotHandler) Affixesall(s *discordgo.Session, m *discordgo.MessageCreate) {
	var message = ""
	for _, aff := range handler.battlenet.Affixes_map {
		message += fmt.Sprintf("%s - %s", aff.Name, aff.Description)

	}
	s.ChannelMessageSend(m.ChannelID, message)
}
