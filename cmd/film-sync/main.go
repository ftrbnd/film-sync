package main

import (
	"github.com/ftrbnd/film-sync/internal/app"
	"github.com/ftrbnd/film-sync/internal/util"
)

func main() {
	err := app.Bootstrap()
	if err != nil {
		util.CheckError("Error found", err)
	}
}
