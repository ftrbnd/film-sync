package files

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func DownloadFrom(link string) string {
	log.Default().Printf("Downloading from %s...", link)

	wd, err := os.Getwd()
	util.CheckError("Failed to get working directory", err)

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
	util.CheckError("Failed to rename file", err)
	log.Default().Println("Renamed file to", newName)

	return newName
}
