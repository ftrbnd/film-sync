package app

import (
	"log"
	"time"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/gmail"
	"github.com/ftrbnd/film-sync/internal/server"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	gmailAPI "google.golang.org/api/gmail/v1"
)

func scheduleJob(c *mongo.Client, s *gmailAPI.Service) {
	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				newLinks := gmail.CheckEmail(c, s)
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

	client := database.Connect()
	service := gmail.GetGmailService()
	scheduleJob(client, service)
	server.Listen()
}
