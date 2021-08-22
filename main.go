package main

import "fmt"

//github.com/PuerkitoBio/goquery html

func main() {
	bot, err := initBot()
	if err != nil {
		fmt.Println("Can't init opb bot ", err)
	}
	bot.Start()
}
