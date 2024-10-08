package files

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func DownloadFrom(link string) (string, error) {
	log.Default().Printf("Downloading from %s...", link)

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %v", err)
	}

	u := launcher.New().Leakless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(link)

	page.MustElementR("button", "Accept All").MustClick()
	log.Default().Println("Clicked on 'Accept All'")

	page.MustElementR("button", "I agree").MustClick()
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

	page.MustElementR("button", "Download").MustClick()
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
