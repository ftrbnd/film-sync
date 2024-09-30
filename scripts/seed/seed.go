package main

import (
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
)

func main() {
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
	err = google.GmailService(authCodeReceived)
	if err != nil {
		panic(err)
	}

	_, err = google.CheckEmail()
	if err != nil {
		panic(err)
	}

	<-authCodeReceived
}
