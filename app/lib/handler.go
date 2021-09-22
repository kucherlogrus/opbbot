package lib

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"opb_bot/lib/battlenet"
	"opb_bot/lib/db"
	"opb_bot/lib/raiderio"
	"reflect"
	"regexp"
	"strings"
)

var (
	server_id             = "701570658919907396"
	main_channel_id       = "701570660014620764"
	test_channel_id       = "879021343981596732"
	free_games_channel_id = "710519845799723090"
	news_channe_id        = "882370165159903283"
)

type BotHandler struct {
	db_instance   *db.DBHandler
	Methods       map[string]reflect.Method
	raider        *raiderio.RaiderApi
	battlenet     *battlenet.Battlenet
	command_regex *regexp.Regexp
}

func getCommand(regex *regexp.Regexp, message string) string {
	match := regex.FindStringSubmatch(message)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func InitHandler(b_handler *BotHandler, db_instance *db.DBHandler) error {
	r, _ := regexp.Compile("(^\\/[a-zA-Z]+)")
	b_handler.command_regex = r
	b_handler.db_instance = db_instance
	b_handler.Methods = map[string]reflect.Method{}
	t := reflect.TypeOf(b_handler)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		method_name := strings.ToLower(m.Name)
		command := strings.Replace(method_name, method_name, "/"+method_name, 1)
		b_handler.Methods[command] = m
	}
	fmt.Printf("Commands list ready. Generated %d commands\n", len(b_handler.Methods))
	b_handler.raider = raiderio.CreateApi()
	fmt.Println("raider.io ready")
	b_handler.battlenet = &battlenet.Battlenet{}
	b_handler.battlenet.InitBattlenetApi(b_handler.db_instance)
	fmt.Println("Battle.net ready")
	fmt.Printf("Battle.net affixes count: %d\n", len(b_handler.battlenet.Affixes_map))
	fmt.Printf("Battle.net dungeons count: %d\n", len(b_handler.battlenet.Dungeon_map))
	return nil
}

func isChannelSupport(channdel_id string) bool {
	return channdel_id == test_channel_id || channdel_id == main_channel_id || channdel_id == news_channe_id
}

func HandleIncomingMessage(handler *BotHandler, s *discordgo.Session, m *discordgo.MessageCreate) {
	if !isChannelSupport(m.ChannelID) {
		return
	}
	command := getCommand(handler.command_regex, m.Content)
	if command != "" {
		method, ok := handler.Methods[command]
		if ok {
			in := []reflect.Value{reflect.ValueOf(handler), reflect.ValueOf(s), reflect.ValueOf(m)}
			method.Func.Call(in)
		}
	}

}

func (handler *BotHandler) _getHelpMessage() (message string) {
	message = "**OPB_BOT** Поддерживаемые команды:\n" +
		"**__/help__** - скоманда показывающая это окно.\n" +
		"**__/raider__ {сервер} {имя}** - отображение информации персонажа в raider.io. Если не задан параметр {сервер} поиск проводится на Гордунни. Параметр сервера нужно указывать латиницей.\n" +
		"**__/affix {имя}__** - отображается информация по аффиксу.\n" +
		"**__/affixes__** - отображаются аффиксы на текущей неделе.\n" +
		"**__/affixesall__** - отображаются все аффиксы из battle.net .\n"
	return
}
