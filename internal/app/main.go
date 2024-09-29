package app

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/server"
	"github.com/ftrbnd/film-sync/internal/util"
	"golang.org/x/oauth2"
)

func startJob(links []string, bot *discordgo.Session) error {
	dst := "output"
	format := "tif"

	for _, link := range links {
		z, err := files.DownloadFrom(link)
		if err != nil {
			return err
		}

		files.Unzip(z, dst, format)
		c, err := files.ConvertToPNG(format, dst)
		if err != nil {
			return err
		}

		files.Upload(dst, z, c, bot)
	}

	return nil
}

func scheduleJob(bot *discordgo.Session) error {
	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				newLinks, err := google.CheckEmail()
				if err != nil {
					return
				}
				log.Default().Printf("Found %d new links", len(newLinks))

				if len(newLinks) > 0 {
					err = startJob(newLinks, bot)
					if err != nil {
						return
					}
				}
			}
		}
	}()

	return nil
}

func Bootstrap() error {
	err := util.LoadEnv()
	if err != nil {
		return err
	}

	err = database.Connect()
	if err != nil {
		return err
	}

	bot, err := discord.Session()
	if err != nil {
		return err
	}
	defer bot.Close()

	authCodeReceived := make(chan *oauth2.Token)

	err = google.GmailService(authCodeReceived, bot)
	if err != nil {
		return err
	}
	err = google.DriveService(authCodeReceived, bot)
	if err != nil {
		return err
	}

	go scheduleJob(bot)
	err = server.Listen(authCodeReceived)
	if err != nil {
		return err
	}

	return nil
}
