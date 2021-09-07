package lib

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"opb_bot/lib/db"
	"os"
	"os/signal"
	"syscall"
	"time"
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
		if m.Content == "/job_newsupdate" {
			bot.updateWoWNews()
		}
		if m.Content == "/job_egsupdate" {
			bot.Egsupdates()
		}

	})
	bot.session.Identify.Intents = discordgo.IntentsGuildMessages
	err := bot.session.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}
	bot.startCronJobs()
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	bot.session.Close()
}

func (bot *OPB_Bot) startCronJobs() {
	scheduler := gocron.NewScheduler(time.UTC)
	_, err := scheduler.Cron("*/20 7-18 * * 1-5").Do(bot.updateWoWNews)
	if err != nil {
		fmt.Println("Can't init cron job checkCron")
		return
	}
	_, err = scheduler.Cron("1 12 * * 1-5").Do(bot.Egsupdates)
	if err != nil {
		fmt.Println("Can't init cron job checkCron")
		return
	}
	scheduler.StartAsync()
}
