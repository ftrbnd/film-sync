package app

import (
	"log"
	"time"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/server"
	"github.com/ftrbnd/film-sync/internal/util"
	"golang.org/x/oauth2"
)

func startJob(links []string) error {
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

		s3Url, driveUrl, message, err := files.Upload(dst, z, c)
		if err != nil {
			return err
		}

		err = discord.SendSuccessMessage(s3Url, driveUrl, message)
		if err != nil {
			return err
		}
	}

	return nil
}

func scheduleJob() error {
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
					err = startJob(newLinks)
					if err != nil {
						panic(err)
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
	defer database.Disconnect()

	err = discord.OpenSession()
	if err != nil {
		return err
	}
	defer discord.CloseSession()

	authCodeReceived := make(chan bool)

	count, err := database.TokenCount()
	if err != nil || count == 0 {
		config, _ := google.Config()
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		discord.SendAuthMessage(authURL)
	} else {
		go func() {
			authCodeReceived <- true
			authCodeReceived <- true
		}()
	}

	google.GmailService(authCodeReceived)
	google.DriveService(authCodeReceived)
	err = scheduleJob()
	if err != nil {
		return err
	}

	err = server.Listen(authCodeReceived)
	if err != nil {
		return err
	}

	return nil
}
