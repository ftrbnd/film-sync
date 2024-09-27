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
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

func startJob(links []string, drive *drive.Service, bot *discordgo.Session) {
	dst := "output"
	format := "tif"

	for _, link := range links {
		z := files.DownloadFrom(link)
		files.Unzip(z, dst, format)
		c := files.ConvertToPNG(format, dst)
		files.Upload(dst, z, c, drive, bot)
	}
}

func scheduleJob(acr chan *oauth2.Token, client *mongo.Client, bot *discordgo.Session) {
	gmail := google.GmailService(acr, client, bot)
	drive := google.DriveService(acr, client, bot)

	ticker := time.NewTicker(24 * time.Hour)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				newLinks := google.CheckEmail(client, gmail)
				log.Default().Printf("Found %d new links", len(newLinks))

				if len(newLinks) > 0 {
					startJob(newLinks, drive, bot)
				}
			}
		}
	}()
}

func Bootstrap() {
	err := godotenv.Load()
	if err != nil {
		log.Default().Println("Failed to load .env file")
	}

	client := database.Connect()
	bot := discord.Session()
	defer bot.Close()

	authCodeReceived := make(chan *oauth2.Token)

	go scheduleJob(authCodeReceived, client, bot)
	server.Listen(authCodeReceived, client)
}
