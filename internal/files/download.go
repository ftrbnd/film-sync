package files

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
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

// findAndClickLinkOrButton tries a <button> match first, then an <a> with the same text regex (Filemail often uses links).
func findAndClickLinkOrButton(page *rod.Page, label string) error {
	err := findAndClickButton(page, label)
	if err == nil {
		return nil
	}
	err2 := rod.Try(func() {
		page.MustWaitDOMStable()
		el := page.Timeout(10*time.Second).MustElementR("a", label)
		el.Timeout(10 * time.Second).MustClick()
	})
	if errors.Is(err2, context.DeadlineExceeded) {
		return fmt.Errorf("%v: no <a> matching %q either", err, label)
	}
	if err2 != nil {
		return fmt.Errorf("button click failed: %v; link click failed: %w", err, err2)
	}
	log.Default().Printf("Found and clicked <a> matching %q", label)
	return nil
}

// waitDownloadHeartbeat logs periodically while waiting for Rod's WaitDownload to finish.
// Close done when the download wait completes (defer close(done) from the caller).
func waitDownloadHeartbeat(done <-chan struct{}) {
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				log.Default().Println("[Download] still waiting for browser download (CDP) to finish...")
			}
		}
	}()
}

// DownloadFrom picks a provider-specific browser flow from the transfer URL.
func DownloadFrom(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("invalid download URL: %w", err)
	}
	host := strings.ToLower(u.Host)
	if strings.Contains(host, "filemail.com") && strings.HasPrefix(u.Path, "/t/") {
		return downloadFilemail(link)
	}
	return downloadWeTransfer(link)
}
