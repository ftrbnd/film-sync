package app

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

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

		msg := err.Error()
		hasAuthErr := strings.Contains(msg, "service hasn't been initialized") ||
			strings.Contains(msg, "expired") ||
			strings.Contains(msg, "invalid_grant")
		if hasAuthErr {
			config, _ := google.Config()
			authURL := google.AuthURL(config)
			discord.SendAuthMessage(authURL)
			log.Default().Println("[Google] Sent auth request to user via Discord")
		}

		return err
	}

	for _, pending := range emails {
		if err := processPendingEmail(pending); err != nil {
			log.Default().Printf("[Film Sync] Skipping email %s: %v", pending.Message.Id, err)
			sentAt := emailSentAt(pending)
			if isExpiredLinkFailure(pending.Provider, sentAt, err) {
				if markErr := markExpiredLink(pending); markErr != nil {
					log.Default().Printf("[Film Sync] Failed to flag expired email %s: %v", pending.Message.Id, markErr)
				}
			}
			discord.SendEmailJobFailure(pending.Provider, pending.Studio, sentAt, err)
			continue
		}
	}

	return nil
}

func isExpiredLinkFailure(provider string, sentAt time.Time, err error) bool {
	return files.IsTransferExpired(provider, sentAt) || files.LooksLikeExpiredDownloadUI(err)
}

func markExpiredLink(pending google.PendingDownload) error {
	scan := database.FilmScan{
		ID:          bson.NewObjectID(),
		EmailID:     pending.Message.Id,
		Provider:    pending.Provider,
		Studio:      pending.Studio,
		LinkExpired: true,
	}
	_, err := database.AddScan(scan)
	return err
}

func emailSentAt(pending google.PendingDownload) time.Time {
	if pending.Message != nil && pending.Message.InternalDate > 0 {
		return time.UnixMilli(pending.Message.InternalDate)
	}
	return time.Time{}
}

func processPendingEmail(pending google.PendingDownload) error {
	url, err := google.GetDownloadURL(pending.Message)
	if err != nil {
		return err
	}

	cldFolder, driveFolderID, message, err := processImages(url)
	if err != nil {
		return err
	}

	newScan := database.FilmScan{
		ID:            bson.NewObjectID(),
		EmailID:       pending.Message.Id,
		Provider:      pending.Provider,
		Studio:        pending.Studio,
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
		// Upload + DB write already succeeded; don't surface this as a download failure.
		log.Default().Printf("[Film Sync] Upload succeeded for email %s but Discord success message failed: %v", pending.Message.Id, err)
	}

	http.SendDeployRequest(message)
	return nil
}

func processImages(downloadURL string) (string, string, string, error) {
	dst := "output"
	format := "tif"

	log.Default().Println("Running garbage collection...")
	runtime.GC()

	z, err := files.DownloadFrom(downloadURL)
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
