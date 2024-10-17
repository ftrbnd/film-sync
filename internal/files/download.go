package files

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var browser *rod.Browser

func StartBrowser() error {
	path, _ := launcher.LookPath()
	u, err := launcher.New().Bin(path).Leakless(false).Headless(true).NoSandbox(true).Set("disable-gpu").RemoteDebuggingPort(9222).Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %v", err)
	}
	browser = rod.New().ControlURL(u)

	err = browser.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to browser: %v", err)
	}

	log.Default().Println("[Files] Browser ready")
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

	button, err := page.ElementR("button", "Accept All")
	if err != nil {
		return "", fmt.Errorf("failed to find Accept All button: %v", err)
	}

	err = button.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return "", fmt.Errorf("failed to click on Accept All button: %v", err)
	}
	log.Default().Println("Clicked on 'Accept All'")

	button, err = page.ElementR("button", "I agree")
	if err != nil {
		return "", fmt.Errorf("failed to find I Agree button: %v", err)
	}

	err = button.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return "", fmt.Errorf("failed to click on I Agree button: %v", err)
	}
	log.Default().Println("Clicked on 'I agree'")

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

	button, err = page.ElementR("button", "Download")
	if err != nil {
		return "", fmt.Errorf("failed to find Download button: %v", err)
	}

	err = button.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return "", fmt.Errorf("failed to click on Download button: %v", err)
	}
	log.Default().Println("Clicked on 'Download'")

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
