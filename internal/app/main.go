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
	"google.golang.org/api/drive/v3"
)

func startJob(links []string, drive *drive.Service, bot *discordgo.Session) error {
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

		files.Upload(dst, z, c, drive, bot)
	}

	return nil
}

func scheduleJob(acr chan *oauth2.Token, bot *discordgo.Session) error {
	gmail, err := google.GmailService(acr, bot)
	if err != nil {
		return err
	}
	drive, err := google.DriveService(acr, bot)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				newLinks, err := google.CheckEmail(gmail)
				if err != nil {
					return
				}
				log.Default().Printf("Found %d new links", len(newLinks))

				if len(newLinks) > 0 {
					err = startJob(newLinks, drive, bot)
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

	go scheduleJob(authCodeReceived, bot)
	err = server.Listen(authCodeReceived)
	if err != nil {
		return err
	}

	return nil
}
