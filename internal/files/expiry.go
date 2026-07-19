package files

import (
	"fmt"
	"strings"
	"time"
)

// ProviderDisplayName maps a From address to a short provider label for UI.
func ProviderDisplayName(provider string) string {
	p := strings.ToLower(provider)
	switch {
	case strings.Contains(p, "filemail"):
		return "Filemail"
	case strings.Contains(p, "wetransfer") || strings.Contains(p, "we.tl"):
		return "WeTransfer"
	default:
		return provider
	}
}

// StudioDisplayName maps a studio Reply-To address to a short label for UI.
func StudioDisplayName(studio string) string {
	s := strings.ToLower(strings.TrimSpace(studio))
	switch {
	case strings.Contains(s, "tomphotostudio"):
		return "Tom's Photo Studio"
	case strings.Contains(s, "goldphoto"):
		return "Gold Photo"
	default:
		return studio
	}
}

// TransferSourceLabel formats "Studio via Provider" for Discord and logs.
func TransferSourceLabel(provider, studio string) string {
	prov := ProviderDisplayName(provider)
	stud := StudioDisplayName(studio)
	if stud == "" {
		return prov
	}
	if prov == "" {
		return stud
	}
	return stud + " via " + prov
}

// IsTransferExpired reports whether a transfer link is past the provider's lifetime
// (Filemail: 1 month, WeTransfer: 5 days) based on when the email was sent.
func IsTransferExpired(provider string, sentAt time.Time) bool {
	return TransferExpiryMessage(provider, sentAt) != ""
}

// TransferExpiryMessage returns a short note when the email is older than the
// provider's typical link lifetime; otherwise empty.
func TransferExpiryMessage(provider string, sentAt time.Time) string {
	if sentAt.IsZero() {
		return ""
	}

	now := time.Now()
	ageDays := int(now.Sub(sentAt).Hours() / 24)
	p := strings.ToLower(provider)

	switch {
	case strings.Contains(p, "filemail"):
		if now.After(sentAt.AddDate(0, 1, 0)) {
			return fmt.Sprintf("Filemail links expire after 1 month; this email is %d days old.", ageDays)
		}
	case strings.Contains(p, "wetransfer") || strings.Contains(p, "we.tl"):
		if now.After(sentAt.Add(5 * 24 * time.Hour)) {
			return fmt.Sprintf("WeTransfer links expire after 5 days; this email is %d days old.", ageDays)
		}
	}
	return ""
}

// LooksLikeExpiredDownloadUI is true for WeTransfer-style failures where the
// Download control never appears (common when the link has expired).
func LooksLikeExpiredDownloadUI(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "timed out finding 'Download' button") ||
		strings.Contains(msg, "'Download' button was not found")
}
