package app

import (
	"context"
	"log"

	"github.com/ftrbnd/film-sync/internal/cloudinary"
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/http"
	"github.com/ftrbnd/film-sync/internal/util"
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

	err = cloudinary.SetCredentials()
	if err != nil {
		return err
	}

	ctx := context.Background()
	config, err := google.Config()
	if err != nil {
		return err
	}

	err = google.StartServices(ctx, config)
	if err != nil {
		authURL := google.AuthURL(config)
		discord.SendAuthMessage(authURL)
		log.Default().Println("[Google] Sent auth request to user via Discord")
	}

	err = files.StartBrowser()
	if err != nil {
		return err
	}

	err = http.Listen(ctx, config, checkEmail)
	if err != nil {
		log.Default().Printf("error starting server: %v", err)
	}

	return nil
}
