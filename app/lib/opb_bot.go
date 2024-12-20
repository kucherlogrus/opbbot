package lib

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"net/http"
	"opb_bot/lib/db"
	"opb_bot/lib/gpt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type OPB_Bot struct {
	session         *discordgo.Session
	db              *db.DBHandler
	handler         *BotHandler
	chat_gpt_client *gpt.OpenaiApiClient
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
	chat_gpt_api_key := os.Getenv("OPENAI_API_KEY")
	chat_gpt_client := gpt.InitOpenaiApiClient(chat_gpt_api_key)
	bot = &OPB_Bot{session, dbHandler, bot_handler, chat_gpt_client}
	return
}

func (bot *OPB_Bot) Start() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		HandleIncomingMessage(bot.handler, s, m)
		if m.Content == "/job_newsupdate" {
			bot.updateWoWNews()
		}
		if m.Content == "/job_egsupdate" {
			bot.Egsupdates()
		}
		if m.Content == "/bot_exit" {
			if m.ChannelID == test_channel_id {
				sc <- os.Kill
			}
		}
		if strings.HasPrefix(m.Content, "/wow_new_parse") {
			if m.ChannelID == test_channel_id {
				bot.parseWoWNew(m.Content)
			}
		}
		if m.Content == "/battle_net_reload" {
			if m.ChannelID == test_channel_id {
				bot.handler.ReloadBattleNetDungeons(bot.session, m)
			}
		}
	})
	bot.session.Identify.Intents = discordgo.IntentsGuildMessages
	err := bot.session.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}
	bot.startCronJobs()
	http.HandleFunc("/version_update", bot.gitHook)
	http_server := &http.Server{Addr: ":8080", Handler: nil}
	go func() {
		http_server.ListenAndServe()
	}()
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	<-sc
	bot.session.Close()
	http_server.Close()
}

func (bot *OPB_Bot) startCronJobs() {
	location, err := time.LoadLocation("Europe/Kiev")
	if err != nil {
		panic("Can't gen location time. ")
	}

	scheduler := gocron.NewScheduler(location)
	_, err = scheduler.Cron("*/10 10-21 * * 1-5").Do(bot.updateWoWNews)
	if err != nil {
		fmt.Println("Can't init cron job checkCron")
		return
	}
	_, err = scheduler.Cron("10 9,12,15,18,21 * * *").Do(bot.Egsupdates)
	if err != nil {
		fmt.Println("Can't init cron job checkCron")
		return
	}

	_, err = scheduler.Cron("15 10 * * 3").Do(bot.updateAffixes)
	if err != nil {
		fmt.Println("Can't init cron job checkCron")
		return
	}

	scheduler.StartAsync()
}

func (bot *OPB_Bot) gitHook(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	message := r.Form.Get("message")
	bot_msg := "Bot updated to new version: \n-----------------------------\n"
	message = bot_msg + message
	message = message + "\n-----------------------------"
	fmt.Println(message)
	bot.newVersionBotNotification(message)
	fmt.Fprintf(w, "OK")
}
