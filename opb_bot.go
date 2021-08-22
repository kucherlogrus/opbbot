package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
)

import db "opb_bot/lib/db"
import egs "opb_bot/lib/crawlers"

var (
	server_id             = "701570658919907396"
	test_channel_id       = "879021343981596732"
	free_games_channel_id = "710519845799723090"
)

type OPB_Bot struct {
	Server_id string

	Free_games_channel_id string

	Main_channel_id string

	session *discordgo.Session

	db *db.DBHandler
}

func initBot() (bot *OPB_Bot, err error) {
	bot_token := os.Getenv("opb_bot_token")
	if bot_token == "" {
		err = fmt.Errorf("Variable opb_bot_token is empty")
		return
	}

	session, err := discordgo.New("Bot " + bot_token)

	if err != nil {
		fmt.Println("Can't create discord bot ", err)
		return
	}

	dbHandler := &db.DBHandler{}
	err = dbHandler.Init()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	bot = &OPB_Bot{server_id, free_games_channel_id, test_channel_id, session, dbHandler}
	return

}

func (bot *OPB_Bot) Start() {
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.Content == "/help" {
			text_message := bot.getHelpMessage()
			s.ChannelMessageSend(m.ChannelID, text_message)
		}

		if m.Content == "ping" && m.ChannelID != bot.Free_games_channel_id {
			s.ChannelMessageSend(m.ChannelID, "Pong!")
		}

		if m.Content == "pong" && m.ChannelID != bot.Free_games_channel_id {
			s.ChannelMessageSend(m.ChannelID, "Ping!")
		}

		if m.Content == "/egsupdate" {
			games, err := egs.ParseFreeEgsGamesUrls()
			if err != nil {
				fmt.Println("Error ParseFreeEgsGamesUrls,", err)
			}
			for _, game := range games {
				game_string := fmt.Sprintf("Название игры: %s\nОписание: %s\nURL: %s\n\n\n", game.Title, game.Description, game.URL)
				s.ChannelMessageSend(free_games_channel_id, game_string)
			}

		}
	})

	bot.session.Identify.Intents = discordgo.IntentsGuildMessages

	err := bot.session.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	bot.session.Close()
}

func (bot *OPB_Bot) getHelpMessage() (message string) {
	message = "Поддерживаемые команды:\n" +
		"**__/help__** - скоманда показывающая это окно.\n" +
		"**__/egsupdate__** - обновление списка бесплатных игр в Epic Games Store"
	return
}
