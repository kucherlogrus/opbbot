package lib

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"opb_bot/lib/egs"
	"regexp"
	"strings"
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
	messages_raw, _ := s.ChannelMessages(free_games_channel_id, 0, m.ID, "", "")
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
			s.ChannelMessageDelete(free_games_channel_id, value)
		}
	}
	//add new games
	for key, game := range current_free_games {
		_, exists := chat_free_games[key]
		if !exists {
			game_string := fmt.Sprintf("\n#id:%s\nНазвание игры: %s\nОписание: %s\nURL: %s\n", game.ID, game.Title, game.Description, game.URL)
			s.ChannelMessageSend(free_games_channel_id, game_string)
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
	for _, aff := range affixes {
		text_affixes = append(text_affixes, fmt.Sprintf("%s : %s", aff.Name, aff.Description))
	}
	message := strings.Join(text_affixes, "\n\n")
	s.ChannelMessageSend(test_channel_id, message)
}
