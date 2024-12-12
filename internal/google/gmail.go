package google

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/util"
	"google.golang.org/api/gmail/v1"
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

	return res.Messages, nil
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
			if header.Name == fieldName {
				if header.Value == fieldValue {
					emails = append(emails, msg)
				}
			}
		}
	}

	return emails, nil
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

	data := msg.Payload.Parts[0].Body.Data
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("unable to decode message body: %v", err)
	}

	lines := strings.Split(string(decoded), "\n")
	url := lines[6] // or find by index of https://wetransfter.com/downloads

	return url, nil
}

func fetchEmails() ([]*gmail.Message, error) {
	err := checkGmailService()
	if err != nil {
		return nil, err
	}

	fromEmail, err := util.LoadEnvVar("FROM_EMAIL")
	if err != nil {
		return nil, err
	}
	replyToEmail, err := util.LoadEnvVar("REPLY_TO_EMAIL")
	if err != nil {
		return nil, err
	}

	messages, err := getEmailsBySender(fromEmail)
	if err != nil {
		return nil, err
	}
	filtered, err := filterEmailsByMetadata(messages, "Reply-To", replyToEmail)
	if err != nil {
		return nil, err
	}

	return filtered, nil
}

func CheckForNewEmails() ([]*gmail.Message, error) {
	log.Default().Println("[Google] Checking email...")

	err := checkGmailService()
	if err != nil {
		return nil, err
	}

	emails, err := fetchEmails()
	if err != nil {
		return nil, err
	}
	scans, err := database.GetScans(false)
	if err != nil {
		return nil, err
	}

	var newEmails []*gmail.Message

	for _, email := range emails {
		exists := database.EmailExists(scans, email)
		if !exists {
			newEmails = append(newEmails, email)
		}
	}

	return newEmails, nil
}

func checkGmailService() error {
	if gmailSrv == nil {
		return fmt.Errorf("gmail service hasn't been initialized")
	}

	return nil
}
