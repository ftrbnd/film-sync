package files

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
)

func downloadWeTransfer(link string) (string, error) {
	log.Default().Println("Starting WeTransfer download... Link:", link)

	if browser == nil {
		return "", errors.New("browser has not been started")
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %v", err)
	}

	page := browser.MustPage()
	err = rod.Try(func() {
		page.MustWaitDOMStable()
		page.Timeout(10 * time.Second).MustNavigate(link)
		log.Default().Println("Successfully navigated to page")
	})
	if errors.Is(err, context.DeadlineExceeded) {
		return "", fmt.Errorf("timed out in navigating to link: %v", err)
	} else if err != nil {
		return "", fmt.Errorf("failed to navigate to link: %v", err)
	}

	btnText := "Accept All"
	err = findAndClickButton(page, btnText)
	if err != nil {
		log.Default().Printf("'%s' button was not found", btnText)
	}
	btnText = "I agree"
	err = findAndClickButton(page, btnText)
	if err != nil {
		log.Default().Printf("'%s' button was not found", btnText)
	}

	page.MustWaitDOMStable()

	wait, cancel := waitBrowserDownload(page.Browser().Timeout(45*time.Minute), wd)
	defer cancel()

	btnText = "Download"
	err = findAndClickButton(page, btnText)
	if err != nil {
		return "", fmt.Errorf("'%s' button was not found: %v", btnText, err)
	}

	log.Default().Println("[Download] click finished; waiting for Chrome to save the file...")
	info, err := wait()
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", fmt.Errorf("download wait timed out or was cancelled before the file finished")
	}

	file := filepath.Join(wd, info.GUID)
	log.Default().Println("Saved", file)

	newName := filepath.Join(wd, info.SuggestedFilename)
	err = os.Rename(file, newName)
	if err != nil {
		return "", fmt.Errorf("failed to rename file: %v", err)
	}

	log.Default().Println("Renamed file to", newName)
	return newName, nil
}
