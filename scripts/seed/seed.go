package main

import (
	"context"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	err := godotenv.Load()
	util.CheckError("Error loading .env file", err)

	client := database.Connect()
	defer client.Disconnect(context.Background())

	bot := discord.Session()
	defer bot.Close()

	authCodeReceived := make(chan *oauth2.Token)
	service := google.GmailService(authCodeReceived, client, bot)

	google.CheckEmail(client, service)

	<-authCodeReceived
}
