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

func downloadFilemail(link string) (string, error) {
	log.Default().Println("Starting Filemail download... Link:", link)

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
		page.Timeout(30 * time.Second).MustNavigate(link)
		log.Default().Println("Successfully navigated to Filemail page")
	})
	if errors.Is(err, context.DeadlineExceeded) {
		return "", fmt.Errorf("timed out in navigating to link: %v", err)
	} else if err != nil {
		return "", fmt.Errorf("failed to navigate to link: %v", err)
	}

	page.MustWaitDOMStable()

	// Cookie / consent banners (best-effort; ignore failures).
	for _, label := range []string{"(?i)accept all", "(?i)^accept$", "(?i)agree"} {
		_ = findAndClickButton(page, label)
	}

	page.MustWaitDOMStable()

	wait, cancel := waitBrowserDownload(page.Browser().Timeout(45*time.Minute), wd)
	defer cancel()

	clicked := false
	for _, label := range []string{"Download all files", "Download all", "(?i)download"} {
		err = findAndClickLinkOrButton(page, label)
		if err == nil {
			clicked = true
			break
		}
		log.Default().Printf("[Download] Filemail control %q not usable: %v", label, err)
	}
	if !clicked {
		return "", fmt.Errorf("download control: could not find a Filemail download button/link")
	}

	log.Default().Println("[Download] click finished; waiting for Chrome to save the file...")
	info, err := wait()
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", fmt.Errorf("download wait timed out or was cancelled before the file finished (wait returned nil)")
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
