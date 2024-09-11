package app

import (
	"log"
	"time"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/gmail"
	"github.com/ftrbnd/film-sync/internal/server"
	"github.com/joho/godotenv"
)

func scheduleJob() {
	client := database.Connect()
	service := gmail.GetGmailService()

	ticker := time.NewTicker(5 * time.Second)
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
					// TODO: open links and download
				}
			}
		}
	}()
}

func Bootstrap() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	go scheduleJob()
	server.Listen()
}
