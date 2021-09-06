package main

import (
	"fmt"
	lb "opb_bot/lib"
)

func main() {
	bot, err := lb.InitBot()
	if err != nil {
		fmt.Println("Can't init opb bot ", err)
		return
	}
	bot.Start()
}
