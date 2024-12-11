package app

import (
	"fmt"
	"log"

	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
)

func runJob(links []string) error {
	dst := "output"
	format := "tif"

	for _, link := range links {
		z, err := files.DownloadFrom(link)
		if err != nil {
			return fmt.Errorf("failed to download from link: %v", err)
		}

		files.Unzip(z, dst, format)
		c, err := files.ConvertToPNG(format, dst)
		if err != nil {
			return fmt.Errorf("failed to convert to png: %v", err)
		}

		cldFolder, driveFolderID, message, err := files.Upload(dst, z, c)
		if err != nil {
			return fmt.Errorf("failed to upload files: %v", err)
		}

		err = discord.SendSuccessMessage(cldFolder, driveFolderID, message)
		if err != nil {
			return fmt.Errorf("failed to send discord success message: %v", err)
		}
	}

	log.Default().Println("[Film Sync] Finished running daily job!")
	return nil
}
