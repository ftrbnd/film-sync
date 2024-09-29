package main

import (
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
	"golang.org/x/oauth2"
)

func main() {
	err := util.LoadEnv()
	if err != nil {
		panic(err)
	}

	err = database.Connect()
	if err != nil {
		panic(err)
	}

	bot, err := discord.Session()
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	authCodeReceived := make(chan *oauth2.Token)
	service, err := google.GmailService(authCodeReceived, bot)
	if err != nil {
		panic(err)
	}

	google.CheckEmail(service)

	<-authCodeReceived
}
