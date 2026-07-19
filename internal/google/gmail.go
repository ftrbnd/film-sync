package google

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/mail"
	"regexp"
	"strings"

	"github.com/ftrbnd/film-sync/internal/config"
	"github.com/ftrbnd/film-sync/internal/database"
	"google.golang.org/api/gmail/v1"
)

// PendingDownload is a new Gmail message plus the provider/studio pairing that matched it.
type PendingDownload struct {
	Message  *gmail.Message
	Provider string
	Studio   string
}

// filemail first: avoids picking a generic link if both appear in HTML.
var (
	reFilemail    = regexp.MustCompile(`https://[\w.-]+\.filemail\.com/t/[\w-]+`)
	reWeTransfer  = regexp.MustCompile(`https://(?:[\w.-]+\.)?(?:wetransfer\.com|we\.tl)[^\s"'<>)\]]*`)
)

func getEmailsBySender(sender string) ([]*gmail.Message, error) {
	err := checkGmailService()
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("from:%s", sender)

	res, err := gmailSrv.Users.Messages.List("me").Q(q).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages from %s: %v", sender, err)
	}

	if res.Messages == nil {
		return nil, nil
	}
	return res.Messages, nil
}

func headerValuesEqual(fieldName, headerValue, want string) bool {
	// Reply-To may include a display name; compare canonical addresses when possible.
	if strings.EqualFold(strings.TrimSpace(fieldName), "Reply-To") {
		hAddr, err1 := mail.ParseAddress(strings.TrimSpace(headerValue))
		wAddr, err2 := mail.ParseAddress(strings.TrimSpace(want))
		if err1 == nil && err2 == nil {
			return strings.EqualFold(hAddr.Address, wAddr.Address)
		}
	}
	return strings.TrimSpace(headerValue) == strings.TrimSpace(want)
}

func filterEmailsByMetadata(messages []*gmail.Message, fieldName string, fieldValue string) ([]*gmail.Message, error) {
	err := checkGmailService()
	if err != nil {
		return nil, err
	}

	var emails []*gmail.Message

	for _, msg := range messages {
		message, err := gmailSrv.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve message_%s: %v", msg.Id, err)
		}

		for _, header := range message.Payload.Headers {
			if strings.EqualFold(header.Name, fieldName) && headerValuesEqual(fieldName, header.Value, fieldValue) {
				emails = append(emails, msg)
				break
			}
		}
	}

	return emails, nil
}

func fetchEmails() ([]PendingDownload, error) {
	err := checkGmailService()
	if err != nil {
		return nil, err
	}

	pairings, err := config.LoadProviderStudioPairings()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]PendingDownload)
	for _, p := range pairings {
		messages, err := getEmailsBySender(p.Provider)
		if err != nil {
			return nil, err
		}
		filtered, err := filterEmailsByMetadata(messages, "Reply-To", p.Studio)
		if err != nil {
			return nil, err
		}
		for _, msg := range filtered {
			if msg == nil || msg.Id == "" {
				continue
			}
			if _, ok := seen[msg.Id]; ok {
				continue
			}
			seen[msg.Id] = PendingDownload{
				Message:  msg,
				Provider: p.Provider,
				Studio:   p.Studio,
			}
		}
	}

	out := make([]PendingDownload, 0, len(seen))
	for _, v := range seen {
		out = append(out, v)
	}
	return out, nil
}

func collectDecodedBodies(part *gmail.MessagePart) [][]byte {
	if part == nil {
		return nil
	}
	var out [][]byte
	mt := strings.ToLower(part.MimeType)
	if (strings.HasPrefix(mt, "text/plain") || strings.HasPrefix(mt, "text/html")) && part.Body != nil && part.Body.Data != "" {
		raw := part.Body.Data
		decoded, err := base64.RawURLEncoding.DecodeString(raw)
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(raw)
		}
		if err != nil {
			decoded, _ = base64.StdEncoding.DecodeString(raw)
		}
		if len(decoded) > 0 {
			out = append(out, decoded)
		}
	}
	for _, sub := range part.Parts {
		out = append(out, collectDecodedBodies(sub)...)
	}
	return out
}

func extractDownloadURLFromMessage(msg *gmail.Message) (string, error) {
	if msg.Payload == nil {
		return "", fmt.Errorf("message has no payload")
	}
	blobs := collectDecodedBodies(msg.Payload)
	if len(blobs) == 0 {
		return "", fmt.Errorf("no decodable text/html or text/plain body in message")
	}
	combined := string(bytesJoin(blobs))
	return pickTransferURL(combined)
}

func bytesJoin(bs [][]byte) []byte {
	n := 0
	for _, b := range bs {
		n += len(b)
	}
	out := make([]byte, 0, n)
	for _, b := range bs {
		out = append(out, b...)
	}
	return out
}

// pickTransferURL prefers Filemail /t/ links, then WeTransfer / we.tl.
func pickTransferURL(body string) (string, error) {
	if m := reFilemail.FindString(body); m != "" {
		return m, nil
	}
	if m := reWeTransfer.FindString(body); m != "" {
		return strings.TrimRight(m, ".,;)]\""), nil
	}
	return "", fmt.Errorf("no WeTransfer or Filemail download URL found in message body")
}

func GetDownloadURL(message *gmail.Message) (string, error) {
	err := checkGmailService()
	if err != nil {
		return "", err
	}

	msg, err := gmailSrv.Users.Messages.Get("me", message.Id).Format("full").Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve message: %v", err)
	}

	return extractDownloadURLFromMessage(msg)
}

func CheckForNewEmails() ([]PendingDownload, error) {
	log.Default().Println("[Google] Checking email...")

	err := checkGmailService()
	if err != nil {
		return nil, err
	}

	emails, err := fetchEmails()
	if err != nil {
		return nil, err
	}
	scans, err := database.GetScans()
	if err != nil {
		return nil, err
	}

	var newEmails []PendingDownload

	for _, pending := range emails {
		exists := database.EmailExists(scans, pending.Message)
		if !exists {
			newEmails = append(newEmails, pending)
		}
	}

	log.Default().Printf("[Google] Found %d new emails", len(newEmails))
	return newEmails, nil
}

func checkGmailService() error {
	if gmailSrv == nil {
		return fmt.Errorf("gmail service hasn't been initialized")
	}

	return nil
}
