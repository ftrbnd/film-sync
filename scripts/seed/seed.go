package main

import (
	"context"

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

	client, err := database.Connect()
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	bot, err := discord.Session()
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	authCodeReceived := make(chan *oauth2.Token)
	service, err := google.GmailService(authCodeReceived, client, bot)
	if err != nil {
		panic(err)
	}

	google.CheckEmail(client, service)

	<-authCodeReceived
}
