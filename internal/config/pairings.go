package config

import (
	"fmt"
	"net/mail"
	"os"
	"strings"

	"github.com/ftrbnd/film-sync/internal/util"
)

// ProviderStudioPairing maps a transfer provider From address to the studio Reply-To inbox.
type ProviderStudioPairing struct {
	Provider string // e.g. noreply@wetransfer.com
	Studio   string // e.g. tomphotostudio@gmail.com
}

// ParseFromReplyToMap parses FROM_REPLY_TO_MAP: "from=reply;from2=reply2".
func ParseFromReplyToMap(raw string) ([]ProviderStudioPairing, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("FROM_REPLY_TO_MAP is empty")
	}

	chunks := strings.Split(raw, ";")
	var out []ProviderStudioPairing
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		idx := strings.Index(chunk, "=")
		if idx <= 0 || idx == len(chunk)-1 {
			return nil, fmt.Errorf("invalid pairing %q: expected from=reply", chunk)
		}
		provider := strings.TrimSpace(chunk[:idx])
		studio := strings.TrimSpace(chunk[idx+1:])
		if provider == "" || studio == "" {
			return nil, fmt.Errorf("invalid pairing %q: empty from or reply", chunk)
		}
		if _, err := mail.ParseAddress(provider); err != nil {
			return nil, fmt.Errorf("provider address %q: %w", provider, err)
		}
		if _, err := mail.ParseAddress(studio); err != nil {
			return nil, fmt.Errorf("studio address %q: %w", studio, err)
		}
		out = append(out, ProviderStudioPairing{Provider: provider, Studio: studio})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("FROM_REPLY_TO_MAP produced no pairings")
	}
	return out, nil
}

// LoadProviderStudioPairings reads FROM_REPLY_TO_MAP when set and non-empty;
// otherwise falls back to FROM_EMAIL and REPLY_TO_EMAIL as a single pairing.
func LoadProviderStudioPairings() ([]ProviderStudioPairing, error) {
	raw := strings.TrimSpace(os.Getenv("FROM_REPLY_TO_MAP"))
	if raw != "" {
		return ParseFromReplyToMap(raw)
	}
	from, err := util.LoadEnvVar("FROM_EMAIL")
	if err != nil {
		return nil, err
	}
	replyTo, err := util.LoadEnvVar("REPLY_TO_EMAIL")
	if err != nil {
		return nil, err
	}
	return []ProviderStudioPairing{{Provider: from, Studio: replyTo}}, nil
}
