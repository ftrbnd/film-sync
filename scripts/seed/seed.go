package main

import (
	"github.com/ftrbnd/film-sync/internal/app"
	"github.com/ftrbnd/film-sync/internal/google"
)

func main() {
	err := app.Bootstrap()
	if err != nil {
		panic(err)
	}

	_, err = google.CheckEmail()
	if err != nil {
		panic(err)
	}
}
