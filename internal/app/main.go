package app

import (
	"log"

	"github.com/ftrbnd/film-sync/internal/gmail"
	"github.com/joho/godotenv"
)

func Bootstrap() {
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }

  gmail.ScheduleJob()
}