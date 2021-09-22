package lib

import (
	"fmt"
	"opb_bot/lib/egs"
	"regexp"
	"time"
)

func (bot *OPB_Bot) Egsupdates() {
	fmt.Println("EGS update job start")
	current_free_games, err := egs.ParseFreeEgsGamesUrls()
	if err != nil {
		fmt.Println("Error ParseFreeEgsGamesUrls,", err)
	}

	chat_free_games := map[string]string{}
	r, _ := regexp.Compile("#id:(.*)\n")
	messages_raw, _ := bot.session.ChannelMessages(free_games_channel_id, 0, "", "", "")
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
			fmt.Println("Remove old free game: ", key)
			bot.session.ChannelMessageDelete(free_games_channel_id, value)
		}
	}
	//add new games
	for key, game := range current_free_games {
		_, exists := chat_free_games[key]
		if !exists {
			fmt.Println("Add new free game: ", game.Title)
			game_string := fmt.Sprintf("\n#id:%s\nНазвание игры: %s\nОписание: %s\nURL: %s\n", game.ID, game.Title, game.Description, game.URL)
			bot.session.ChannelMessageSend(free_games_channel_id, game_string)
		}
	}
}

func (bot *OPB_Bot) updateWoWNews() {

	fmt.Println("Update news job start")

	value, err := bot.db.GetActionValue("wownews")
	if err != nil {
		bot.session.ChannelMessageSend(news_channe_id, fmt.Sprintln(err))
		return
	}
	value_t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		bot.session.ChannelMessageSend(news_channe_id, fmt.Sprintln(err))
		return
	}
	news_list, err := bot.handler.battlenet.GetLastNews(value)

	if err != nil {
		bot.session.ChannelMessageSend(news_channe_id, fmt.Sprintln(err))
		return
	}
	count_news := len(news_list)
	if count_news == 0 {
		fmt.Println("No news found.")
		return
	}

	var first_time = news_list[0].Timestr

	first_time_t, err := time.Parse(time.RFC3339, value)

	if first_time_t.Before(value_t) {
		fmt.Println("Last news was already handled.")
		return
	}

	fmt.Println("Last handled new time: ", value)

	var one_time_limit = 5
	var send_count = 0
	for _, new_el := range news_list {
		last_time := new_el.Timestr
		last_time_t, err := time.Parse(time.RFC3339, last_time)
		if err != nil {
			continue
		}
		if last_time_t.Before(value_t) {
			break
		}
		fmt.Println(last_time, value)

		fmt.Printf("Handle title %s: %s\n", new_el.Tittle, last_time)
		new_text, err_n := bot.handler.battlenet.GetNewFromUrl(new_el.URL)
		if err_n != nil {
			fmt.Println(err_n)
			continue
		}

		fmt.Println("len of text: ", len(new_text))

		message := "**__" + new_el.Tittle + "__**\n" + new_text
		max_index := 1999

		index := max_index
		for {
			count := len(message)
			if count <= 2000 {
				bot.session.ChannelMessageSend(news_channe_id, message+"\n")
				bot.session.ChannelMessageSend(news_channe_id, "------------------------------------------------------------------\n")
				break
			}
			for {
				char := message[index]
				if char == ' ' {
					send_message := message[:index]
					message = message[index:]
					bot.session.ChannelMessageSend(news_channe_id, send_message)
					index = max_index
					break
				}
				index--
			}
		}
		send_count += 1
		if send_count >= one_time_limit {
			break
		}
	}
	if first_time_t.After(value_t) {
		bot.db.UpdateActionValue(first_time, "wownews")
	}

}
