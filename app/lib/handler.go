package lib

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"opb_bot/lib/db"
	"opb_bot/lib/raiderio"
	"reflect"
	"strings"
)

var (
	server_id             = "701570658919907396"
	test_channel_id       = "879021343981596732"
	free_games_channel_id = "710519845799723090"
)

type BotHandler struct {
	db_instance *db.DBHandler
	Methods     map[string]reflect.Method
	raider      *raiderio.RaiderApi
}

func InitHandler(b_handler *BotHandler, db_instance *db.DBHandler) {
	b_handler.db_instance = db_instance
	b_handler.Methods = map[string]reflect.Method{}
	t := reflect.TypeOf(b_handler)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		method_name := strings.ToLower(m.Name)
		command := strings.Replace(method_name, method_name, "/"+method_name, 1)
		b_handler.Methods[command] = m
	}

	b_handler.raider = raiderio.CreateApi()

}

func HandleIncomingMessage(handler *BotHandler, s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.ChannelID, m.Content)
	fmt.Println("-----------------------")
	method, ok := handler.Methods[m.Content]
	if ok {
		in := []reflect.Value{reflect.ValueOf(handler), reflect.ValueOf(s), reflect.ValueOf(m)}
		method.Func.Call(in)
	}

}

func (handler *BotHandler) _getHelpMessage() (message string) {
	message = "Поддерживаемые команды:\n" +
		"**__/help__** - скоманда показывающая это окно.\n" +
		"**__/egsupdate__** - обновление списка бесплатных игр в Epic Games Store"
	return
}
