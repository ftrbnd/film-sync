package main

import (
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
	"golang.org/x/oauth2"
)

func init() {
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
}

func main() {
	/**
	- go through each folder in parent folder "Film Photos"
	- for each folder:
		- get all .TIF photos
			- convert to .PNG
			- upload to S3
	*/
	folderID, err := util.LoadEnvVar("DRIVE_FOLDER_ID")
	if err != nil {
		panic(err)
	}

	output := "output"

	err = google.DownloadFolder(folderID, output)
	if err != nil {
		panic(err)
	}

	count, err := files.ConvertToPNG(".tif", output)
	if err != nil {
		panic(err)
	}
	_, _, _, err = files.Upload(output, "output.zip", count)
	if err != nil {
		panic(err)
	}
}
