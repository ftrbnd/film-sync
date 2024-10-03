package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/server"
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

	go func() {
		err = server.Listen(authCodeReceived)
		if err != nil {
			panic(err)
		}
	}()

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
	folderID, err := util.LoadEnvVar("DRIVE_FOLDER_ID")
	if err != nil {
		panic(err)
	}

	output := "output"

	err = google.DownloadFolder(folderID, output)
	if err != nil {
		panic(err)
	}

	count, err := files.ConvertToPNG("tif", output)
	if err != nil {
		panic(err)
	}

	err = filepath.WalkDir(output, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "tif") {
			log.Default().Printf("Removing %s", path)
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}

		return nil
	})

	_, _, _, err = files.Upload(output, "output.zip", count)
	if err != nil {
		panic(err)
	}
}
