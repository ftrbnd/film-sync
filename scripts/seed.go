package main

import (
	"context"
	"log"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/gmail"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func Seed() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := database.Connect()
	defer client.Disconnect(context.Background())

	authCodeReceived := make(chan *oauth2.Token)
	service := gmail.Service(authCodeReceived)

	gmail.CheckEmail(client, service)

	<-authCodeReceived
}
