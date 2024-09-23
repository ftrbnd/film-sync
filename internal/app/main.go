package app

import (
	"log"
	"time"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/gmail"
	"github.com/ftrbnd/film-sync/internal/server"
	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func startJob(links []string) {
	dst := "output"
	initialFormat := "tif"

	for _, link := range links {
		zipFile := files.DownloadFrom(link)
		files.Unzip(zipFile, dst, initialFormat)
		files.ConvertToPNG(initialFormat, dst)
		files.Upload(dst)
	}
}

func scheduleJob(acr chan *oauth2.Token) {
	client := database.Connect()
	service := gmail.Service(acr)

	ticker := time.NewTicker(24 * time.Hour)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				newLinks := gmail.CheckEmail(client, service)
				log.Default().Printf("Found %d new links", len(newLinks))

				if len(newLinks) > 0 {
					startJob(newLinks)
				}
			}
		}
	}()
}

func Bootstrap() {
	err := godotenv.Load()
	util.CheckError("Error loading .env file", err)

	authCodeReceived := make(chan *oauth2.Token)

	go scheduleJob(authCodeReceived)
	server.Listen(authCodeReceived)
}
