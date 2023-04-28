package lib

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"opb_bot/lib/utils"
	"strings"
	"time"
)

func (handler *BotHandler) Help(s *discordgo.Session, m *discordgo.MessageCreate) {
	text_message := handler._getHelpMessage()
	s.ChannelMessageSend(m.ChannelID, text_message)
}

func (handler *BotHandler) Clear(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID == test_channel_id || m.ChannelID == news_channe_id {
		fmt.Println("clear all messages")
		for {
			messages_raw, _ := s.ChannelMessages(m.ChannelID, 0, "", "", "")
			if len(messages_raw) == 0 {
				break
			}
			for _, msg := range messages_raw {
				fmt.Println("clear all message ", msg.ID)
				s.ChannelMessageDelete(m.ChannelID, msg.ID)
			}
		}
	}

}

func (handler *BotHandler) Raider(s *discordgo.Session, m *discordgo.MessageCreate) {
	raw_params := strings.Split(m.Content, " ")
	params := []string{}
	for _, param := range raw_params {
		if param != "" {
			params = append(params, param)
		}
	}
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
	name = strings.Trim(name, " ")
	realm = strings.Trim(realm, " ")
	result, err := handler.raider.GetUserInfo(realm, name)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Ошибка: %s", err))
	}
	var message = ""
	message += fmt.Sprintf("Имя: %s\n", result.Name)
	message += fmt.Sprintf("Сервер: %s\n", result.Realm)
	message += fmt.Sprintf("Уровень экипировки': **__%d__**\n", result.Gear.ItemLevelEquipped)
	message += fmt.Sprintf("Рейтинг:  **__%d__**\n", int(result.MythicPlusScoresBySeason[0].Scores.All))
	message += fmt.Sprintf("Место: мир **__%d__**, сервер **__%d__**\n", result.MythicPlusRanks.Overall.World, result.MythicPlusRanks.Overall.Realm)
	message += fmt.Sprintf("Место по классу: мир **__%d__**, сервер **__%d__**\n", result.MythicPlusRanks.Class.World, result.MythicPlusRanks.Class.Realm)
	message += fmt.Sprintf("Место по фракции: мир **__%d__**, сервер **__%d__**\n", result.MythicPlusRanks.FactionOverall.World, result.MythicPlusRanks.FactionOverall.Realm)
	message += fmt.Sprintf("Место по фракции и классу: мир **__%d__**, сервер **__%d__**\n", result.MythicPlusRanks.FactionClass.World, result.MythicPlusRanks.FactionClass.Realm)
	message += fmt.Sprintf("-----------------------------------\n")
	for _, instance := range result.MythicPlusBestRuns {
		t := instance.CompletedAt
		complete_at := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second())
		tm := instance.ClearTimeMs
		sec := tm / 1000
		msec := tm % 1000
		tm_time := time.Unix(int64(sec), int64(msec*int(time.Millisecond)))
		tm_time_str := fmt.Sprintf("%d мин. %d сек.", tm_time.Minute(), tm_time.Second())
		dungeon_name := handler.battlenet.Dungeon_map[instance.Dungeon]
		time_for_clean := instance.ParTimeMs
		sec = time_for_clean / 1000
		msec = time_for_clean % 1000
		time_for_clean_tm := time.Unix(int64(sec), int64(msec*int(time.Millisecond)))
		message += fmt.Sprintf("%s **__%d__**, пройден %s за %s. Время на прохождение: %d мин\n", dungeon_name.Name, instance.MythicLevel, complete_at, tm_time_str, time_for_clean_tm.Minute())
	}
	message += fmt.Sprintf("-----------------------------------\n")
	r := result.LastCrawledAt
	scan_time := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", r.Day(), r.Month(), r.Year(), r.Hour(), r.Minute(), r.Second())
	message += fmt.Sprintf("Дата обновления raider.io: %s\n", scan_time)
	s.ChannelMessageSend(m.ChannelID, message)
}

func (handler *BotHandler) Affix(s *discordgo.Session, m *discordgo.MessageCreate) {
	params := strings.Split(m.Content, " ")
	params_count := len(params)
	if params_count == 2 {
		affix_name := strings.Title(params[1])
		for _, v := range handler.battlenet.Affixes_map {
			if affix_name == v.Name {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**__%s__** - %s\n", v.Name, v.Description))
				return
			}
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Аффикс с названием '%s' не найден.", params[1]))
		return
	}
	if params_count == 1 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Для выполнения команды необходимо указать название аффикса", params[1]))
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Параметры команды заданы неверно"))
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
	for _, v := range handler.battlenet.Affixes_map {
		next_message := fmt.Sprintf("**__%s__** - %s\n", v.Name, v.Description)
		current_length := len(message)
		next_message_len := len(next_message)
		//Must be 2000 or fewer in length.
		fmt.Println(current_length)
		if next_message_len+current_length > 2000 {
			s.ChannelMessageSend(m.ChannelID, message)
			message = next_message
		} else {
			message += next_message
		}

	}

	send, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		fmt.Println(err)
		return
	}
	utils.PrintType(send)
}
