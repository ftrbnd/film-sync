package files

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	button, err := page.ElementR("button", jsRegex)
	if err != nil {
		return fmt.Errorf("failed to find %s button: %v", jsRegex, err)
	}

	err = button.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return fmt.Errorf("failed to click on %s button: %v", jsRegex, err)
	}

	log.Default().Printf("Clicked on '%s'", jsRegex)
	return nil
}

func DownloadFrom(link string) (string, error) {
	log.Default().Printf("Downloading from %s...", link)

	if browser == nil {
		return "", errors.New("browser has not been started")
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %v", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{
		URL: link,
	})
	if err != nil {
		return "", fmt.Errorf("failed to visit url: %v", err)
	}

	err = findAndClickButton(page, "Accept All")
	if err != nil {
		return "", err
	}

	err = findAndClickButton(page, "I agree")
	if err != nil {
		return "", err
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
