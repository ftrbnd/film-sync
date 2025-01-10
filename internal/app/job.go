package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/http"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func checkEmail() error {
	emails, err := google.CheckForNewEmails()
	if err != nil {
		discord.SendErrorMessage(err)

		hasAuthErr := strings.Contains(err.Error(), "service hasn't been initialized") || strings.Contains(err.Error(), "token expired")
		if hasAuthErr {
			config, _ := google.Config()
			authURL := google.AuthURL(config)
			discord.SendAuthMessage(authURL)
			log.Default().Println("[Google] Sent auth request to user via Discord")
		}

		return err
	}

	for _, email := range emails {
		url, err := google.GetDownloadURL(email)
		if err != nil {
			return err
		}

		cldFolder, driveFolderID, message, err := processImages(url)
		if err != nil {
			discord.SendErrorMessage(err)
			return err
		}

		newScan := database.FilmScan{
			ID:            bson.NewObjectID(),
			EmailID:       email.Id,
			DownloadURL:   url,
			CldFolderName: cldFolder,
			DriveFolderID: driveFolderID,
		}
		_, err = database.AddScan(newScan)
		if err != nil {
			return err
		}

		err = discord.SendSuccessMessage(newScan.ID.Hex(), message)
		if err != nil {
			log.Default().Println(err)
			return fmt.Errorf("failed to send discord success message: %v", err)
		}

		http.SendDeployRequest(message)
	}

	return nil
}

func processImages(weTransferURL string) (string, string, string, error) {
	dst := "output"
	format := "tif"

	z, err := files.DownloadFrom(weTransferURL)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to download from link: %v", err)
	}

	files.Unzip(z, dst, format)
	c, err := files.ConvertToPNG(format, dst)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to convert to png: %v", err)
	}

	cldFolder, driveFolderID, message, err := files.Upload(dst, z, c)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to upload files: %v", err)
	}

	log.Default().Println("[Film Sync] Finished running daily job!")
	return cldFolder, driveFolderID, message, nil
}
