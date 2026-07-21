package files

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
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

type downloadProgress struct {
	mu       sync.Mutex
	started  bool
	filename string
	url      string
	received float64
	total    float64
	state    proto.BrowserDownloadProgressState
}

func (p *downloadProgress) snapshot() (started bool, filename string, received, total float64, state proto.BrowserDownloadProgressState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.started, p.filename, p.received, p.total, p.state
}

func formatDownloadBytes(n float64) string {
	const mb = 1024 * 1024
	if n >= mb {
		return fmt.Sprintf("%.2f MB", n/mb)
	}
	if n >= 1024 {
		return fmt.Sprintf("%.1f KB", n/1024)
	}
	return fmt.Sprintf("%.0f B", n)
}

// downloadInfo is the completed download metadata (Browser CDP events).
type downloadInfo struct {
	GUID              string
	URL               string
	SuggestedFilename string
}

// waitBrowserDownload waits for a Chrome download using Browser.* CDP events
// (Page.download* is deprecated and often never fires on newer Chromium).
// Always defer cancel() so the progress ticker stops if wait() is never called
// (e.g. the download button was not found).
func waitBrowserDownload(b *rod.Browser, dir string) (wait func() (*downloadInfo, error), cancel func()) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	var oldDownloadBehavior proto.BrowserSetDownloadBehavior
	has := b.LoadState("", &oldDownloadBehavior)

	_ = proto.BrowserSetDownloadBehavior{
		Behavior:         proto.BrowserSetDownloadBehaviorBehaviorAllowAndName,
		BrowserContextID: b.BrowserContextID,
		DownloadPath:     absDir,
		EventsEnabled:    true,
	}.Call(b)

	prog := &downloadProgress{}
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				started, filename, received, total, state := prog.snapshot()
				if !started {
					log.Default().Println("[Download] waiting for browser download to begin (CDP)...")
					continue
				}
				if total > 0 {
					pct := received / total * 100
					log.Default().Printf("[Download] %s: %s / %s (%.0f%%) state=%s",
						filename, formatDownloadBytes(received), formatDownloadBytes(total), pct, state)
				} else {
					log.Default().Printf("[Download] %s: %s received (total unknown) state=%s",
						filename, formatDownloadBytes(received), state)
				}
			}
		}
	}()

	var start *proto.BrowserDownloadWillBegin
	var terminalErr error

	waitProgress := b.EachEvent(func(e *proto.BrowserDownloadWillBegin) {
		start = e
		prog.mu.Lock()
		prog.started = true
		prog.filename = e.SuggestedFilename
		prog.url = e.URL
		prog.state = proto.BrowserDownloadProgressStateInProgress
		prog.mu.Unlock()
		log.Default().Printf("[Download] began: file=%q url=%s", e.SuggestedFilename, e.URL)
	}, func(e *proto.BrowserDownloadProgress) bool {
		if start == nil || start.GUID != e.GUID {
			return false
		}
		prog.mu.Lock()
		prog.received = e.ReceivedBytes
		prog.total = e.TotalBytes
		prog.state = e.State
		prog.mu.Unlock()

		switch e.State {
		case proto.BrowserDownloadProgressStateCompleted:
			log.Default().Printf("[Download] completed: %s (%s)",
				start.SuggestedFilename, formatDownloadBytes(e.ReceivedBytes))
			return true
		case proto.BrowserDownloadProgressStateCanceled:
			terminalErr = fmt.Errorf("browser download canceled for %q", start.SuggestedFilename)
			log.Default().Printf("[Download] canceled: %s after %s",
				start.SuggestedFilename, formatDownloadBytes(e.ReceivedBytes))
			return true
		default:
			return false
		}
	})

	var cleanupOnce sync.Once
	cancel = func() {
		cleanupOnce.Do(func() {
			close(done)
			if has {
				_ = oldDownloadBehavior.Call(b)
			} else {
				_ = proto.BrowserSetDownloadBehavior{
					Behavior:         proto.BrowserSetDownloadBehaviorBehaviorDefault,
					BrowserContextID: b.BrowserContextID,
				}.Call(b)
			}
		})
	}

	wait = func() (*downloadInfo, error) {
		defer cancel()
		waitProgress()
		if terminalErr != nil {
			if start == nil {
				return nil, terminalErr
			}
			return &downloadInfo{GUID: start.GUID, URL: start.URL, SuggestedFilename: start.SuggestedFilename}, terminalErr
		}
		if start == nil {
			return nil, fmt.Errorf("download wait finished without a DownloadWillBegin event")
		}
		return &downloadInfo{GUID: start.GUID, URL: start.URL, SuggestedFilename: start.SuggestedFilename}, nil
	}
	return wait, cancel
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
