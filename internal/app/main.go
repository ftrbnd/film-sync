package app

import (
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

	authCodeReceived := make(chan bool)

	count, err := database.TokenCount()
	if err != nil || count == 0 {
		config, _ := google.Config()
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		discord.SendAuthMessage(authURL)
	} else {
		go func() {
			authCodeReceived <- true
			authCodeReceived <- true
		}()
	}

	google.GmailService(authCodeReceived)
	google.DriveService(authCodeReceived)

	err = server.Listen(authCodeReceived)
	if err != nil {
		return err
	}

	return nil
}
