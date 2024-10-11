package app

import (
	"log"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
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

	err = google.StartServices()
	if err != nil {
		config, _ := google.Config()
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		discord.SendAuthMessage(authURL)
		log.Default().Println("[Google] Sent auth request to user via Discord")

	}

	err = server.Listen()
	if err != nil {
		return err
	}

	return nil
}
