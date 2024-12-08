package main

import (
	"context"
	"log"

	"github.com/ftrbnd/film-sync/internal/aws"
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
	defer database.Disconnect()

	err = discord.OpenSession()
	if err != nil {
		panic(err)
	}
	defer discord.CloseSession()

	err = aws.StartClient()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	config, err := google.Config()
	if err != nil {
		panic(err)
	}

	err = google.StartServices(ctx, config)
	if err != nil {
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		discord.SendAuthMessage(authURL)
		log.Default().Println("[Google] Sent auth request to user via Discord")
	}

	_, err = google.CheckEmail()
	if err != nil {
		panic(err)
	}
}
