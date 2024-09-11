package app

import (
	"log"
	"time"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/gmail"
	"github.com/ftrbnd/film-sync/internal/server"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func scheduleJob(c *mongo.Client) {
	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				newLinks := gmail.CheckEmail(c)
				log.Default().Printf("Found %d new links", len(newLinks))
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
	scheduleJob(client)
	server.Listen()
}
