package main

import (
	"context"
	"log"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/gmail"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := database.Connect()
	defer client.Disconnect(context.Background())

	gmail.CheckEmail(client)
}
