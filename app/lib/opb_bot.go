package lib

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"opb_bot/lib/db"
	"os"
	"os/signal"
	"syscall"
)

type OPB_Bot struct {
	session *discordgo.Session
	db      *db.DBHandler
	handler *BotHandler
}

func InitBot() (bot *OPB_Bot, err error) {

	dbHandler := &db.DBHandler{}
	err = dbHandler.Init()
	if err != nil {
		return nil, err
	}
	fmt.Println("Data base initialised")

	discord_access_token, err := dbHandler.GetAccessToken("discord")
	if err != nil {
		fmt.Println("Can't create discord bot, sql error: ", err)
	}

	session, err := discordgo.New("Bot " + discord_access_token)
	if err != nil {
		fmt.Println("Can't create discord bot ", err)
		return
	}
	fmt.Println("Discord bot connected")
	bot_handler := &BotHandler{}
	err = InitHandler(bot_handler, dbHandler)
	if err != nil {
		return nil, err
	}
	fmt.Println("Bot handlers initialised")
	bot = &OPB_Bot{session, dbHandler, bot_handler}
	return
}

func (bot *OPB_Bot) Start() {
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		HandleIncomingMessage(bot.handler, s, m)
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
