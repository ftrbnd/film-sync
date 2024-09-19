package main

import (
	"context"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/gmail"
	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	err := godotenv.Load()
	util.CheckError("Error loading .env file", err)

	client := database.Connect()
	defer client.Disconnect(context.Background())

	authCodeReceived := make(chan *oauth2.Token)
	service := gmail.Service(authCodeReceived)

	gmail.CheckEmail(client, service)

	<-authCodeReceived
}
