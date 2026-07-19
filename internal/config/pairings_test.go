package config

import (
	"os"
	"testing"
)

func TestParseFromReplyToMap(t *testing.T) {
	raw := "noreply@wetransfer.com=tomphotostudio@gmail.com; no-reply@filemail.com = goldphoto@sbcglobal.net "
	got, err := ParseFromReplyToMap(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Provider != "noreply@wetransfer.com" || got[0].Studio != "tomphotostudio@gmail.com" {
		t.Fatalf("pair0: %+v", got[0])
	}
	if got[1].Provider != "no-reply@filemail.com" || got[1].Studio != "goldphoto@sbcglobal.net" {
		t.Fatalf("pair1: %+v", got[1])
	}
}

func TestParseFromReplyToMap_Invalid(t *testing.T) {
	if _, err := ParseFromReplyToMap("foo"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := ParseFromReplyToMap("a=b;nodash"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := ParseFromReplyToMap(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadProviderStudioPairings_FromMap(t *testing.T) {
	t.Setenv("FROM_REPLY_TO_MAP", "a@b.com=c@d.com")
	t.Setenv("FROM_EMAIL", "should@ignore.com")
	t.Setenv("REPLY_TO_EMAIL", "ignore@ignore.com")
	got, err := LoadProviderStudioPairings()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Provider != "a@b.com" || got[0].Studio != "c@d.com" {
		t.Fatalf("%+v", got)
	}
}

func TestLoadProviderStudioPairings_Fallback(t *testing.T) {
	t.Setenv("FROM_REPLY_TO_MAP", "")
	t.Setenv("FROM_EMAIL", "noreply@wetransfer.com")
	t.Setenv("REPLY_TO_EMAIL", "studio@example.com")
	got, err := LoadProviderStudioPairings()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Provider != "noreply@wetransfer.com" || got[0].Studio != "studio@example.com" {
		t.Fatalf("%+v", got)
	}
}

func TestLoadProviderStudioPairings_FallbackMissingFrom(t *testing.T) {
	t.Setenv("FROM_REPLY_TO_MAP", "")
	_ = os.Unsetenv("FROM_EMAIL")
	_ = os.Unsetenv("REPLY_TO_EMAIL")
	_, err := LoadProviderStudioPairings()
	if err == nil {
		t.Fatal("expected error when FROM_EMAIL missing")
	}
}
