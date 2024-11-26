package app

import (
	"context"
	"log"

	"github.com/ftrbnd/film-sync/internal/aws"
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/server"
	"github.com/ftrbnd/film-sync/internal/util"
	"golang.org/x/oauth2"
)

func Bootstrap() error {
	err := util.LoadEnv()
	if err != nil {
		return err
	}

	err = database.Connect()
	if err != nil {
		return err
	}
	defer database.Disconnect()

	err = discord.OpenSession()
	if err != nil {
		return err
	}
	defer discord.CloseSession()

	err = aws.StartClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = google.StartServices(ctx)
	if err != nil {
		config, _ := google.Config()
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		discord.SendAuthMessage(authURL)
		log.Default().Println("[Google] Sent auth request to user via Discord")

	}

	err = files.StartBrowser()
	if err != nil {
		return err
	}

	err = server.Listen(ctx)
	if err != nil {
		log.Default().Printf("error starting server: %v", err)
	}

	return nil
}
