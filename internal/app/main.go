package app

import (
	"log"
	"time"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/server"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func startJob(links []string) {
	dst := "output"
	format := "tif"

	for _, link := range links {
		z := files.DownloadFrom(link)
		files.Unzip(z, dst, format)
		c := files.ConvertToPNG(format, dst)
		files.Upload(dst, z, c)
	}
}

func scheduleJob(acr chan *oauth2.Token) {
	client := database.Connect()
	service := google.Service(acr)

	ticker := time.NewTicker(24 * time.Hour)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				newLinks := google.CheckEmail(client, service)
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
	if err != nil {
		log.Default().Println("Failed to load .env file")
	}

	authCodeReceived := make(chan *oauth2.Token)

	go scheduleJob(authCodeReceived)
	server.Listen(authCodeReceived)
}
