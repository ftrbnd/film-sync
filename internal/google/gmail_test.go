package google

import (
	"strings"
	"testing"
)

func TestPickTransferURL_Filemail(t *testing.T) {
	t.Parallel()
	body := `<a href="https://gold-photo.filemail.com/t/qoVOu41B">click</a>`
	u, err := pickTransferURL(body)
	if err != nil {
		t.Fatal(err)
	}
	if u != "https://gold-photo.filemail.com/t/qoVOu41B" {
		t.Fatalf("got %q", u)
	}
}

func TestPickTransferURL_WeTransfer(t *testing.T) {
	t.Parallel()
	body := "Here is your transfer https://we.tl/t-abc123xyz end"
	u, err := pickTransferURL(body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "we.tl") {
		t.Fatalf("got %q", u)
	}
}

func TestPickTransferURL_WeTransferDownloads(t *testing.T) {
	t.Parallel()
	body := `href="https://wetransfer.com/downloads/aaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/cccccccc"`
	u, err := pickTransferURL(body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "wetransfer.com/downloads/") {
		t.Fatalf("got %q", u)
	}
}

func TestPickTransferURL_FilemailPreferredWhenBoth(t *testing.T) {
	t.Parallel()
	body := `wt https://we.tl/t-zzzz fm https://sub.filemail.com/t/abc`
	u, err := pickTransferURL(body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "filemail.com") {
		t.Fatalf("got %q", u)
	}
}

func TestPickTransferURL_None(t *testing.T) {
	t.Parallel()
	if _, err := pickTransferURL("no links here"); err == nil {
		t.Fatal("expected error")
	}
}
