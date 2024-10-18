package files

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var browser *rod.Browser

func StartBrowser() error {
	var path string

	p, err := util.LoadEnvVar("BROWSER_PATH")
	if err != nil {
		path, _ = launcher.LookPath()
	} else {
		path = p
	}

	u, err := launcher.New().Bin(path).Leakless(false).Headless(true).NoSandbox(true).Set("disable-gpu").RemoteDebuggingPort(9222).Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %v", err)
	}

	browser = rod.New().ControlURL(u)

	err = browser.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to browser: %v", err)
	}

	log.Default().Printf("[Files] Browser ready at %s", path)
	return nil
}

func findAndClickButton(page *rod.Page, jsRegex string) error {
	err := rod.Try(func() {
		page.MustWaitDOMStable()

		button := page.Timeout(5*time.Second).MustElementR("button", jsRegex)
		button.Timeout(5 * time.Second).MustClick()
	})
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("timed out finding '%s' button: %v", jsRegex, err)
	} else if err != nil {
		return fmt.Errorf("failed to find and click '%s' button: %v", jsRegex, err)
	}

	log.Default().Printf("Found and clicked '%s' button", jsRegex)
	return nil
}

func DownloadFrom(link string) (string, error) {
	log.Default().Println("Starting download process...", link)

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

	// sometimes the page will directly go to the Download view, so disregard the next 2 errors
	err = findAndClickButton(page, "Accept All")
	if err != nil {
		log.Default().Println("'Accept All' button was not found")
	}
	err = findAndClickButton(page, "I agree")
	if err != nil {
		log.Default().Println("'I agree' button was not found")
	}

	page.MustWaitDOMStable()

	wait := page.Browser().WaitDownload(wd)

	go browser.EachEvent(func(e *proto.PageDownloadProgress) bool {
		completed := "(unknown)"
		if e.TotalBytes != 0 {
			completed = fmt.Sprintf("%0.2f%%", e.ReceivedBytes/e.TotalBytes*100.0)
		}
		log.Printf("Downloading... %s\n", completed)
		return e.State == proto.PageDownloadProgressStateCompleted
	})()

	err = findAndClickButton(page, "Download")
	if err != nil {
		return "", err
	}

	res := wait()

	file := filepath.Join(wd, res.GUID)
	log.Default().Println("Saved", file)

	newName := filepath.Join(wd, res.SuggestedFilename)
	err = os.Rename(file, newName)
	if err != nil {
		return "", fmt.Errorf("failed to rename file: %v", err)
	}

	log.Default().Println("Renamed file to", newName)
	return newName, nil
}
