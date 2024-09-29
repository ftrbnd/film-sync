package google

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/api/gmail/v1"
)

func getEmailsBySender(sender string) ([]*gmail.Message, error) {
	q := fmt.Sprintf("from:%s", sender)

	res, err := gmailSrv.Users.Messages.List("me").Q(q).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages from %s: %v", sender, err)
	}

	return res.Messages, nil
}

func filterEmailsByMetadata(messages []*gmail.Message, fieldName string, fieldValue string) ([]*gmail.Message, error) {
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

func GetDownloadLink(message *gmail.Message) (string, error) {
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
	link := lines[6] // or find by index of https://wetransfter.com/downloads

	return link, nil
}

func FetchEmails() ([]*gmail.Message, error) {
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

func CheckEmail() ([]string, error) {
	log.Default().Println("Checking email...")

	emails, err := FetchEmails()
	if err != nil {
		return nil, err
	}
	saved, err := database.GetEmails()
	if err != nil {
		return nil, err
	}

	var newLinks []string

	for _, email := range emails {
		exists := database.EmailExists(saved, email)
		if !exists {
			link, err := GetDownloadLink(email)
			if err != nil {
				return nil, err
			}

			newEmail := database.Email{ID: bson.NewObjectID(), EmailID: email.Id, DownloadLink: link}

			database.AddEmail(newEmail)

			newLinks = append(newLinks, link)
		} else {
			log.Default().Printf("Email #%s already exists in database", email.Id)
		}
	}

	return newLinks, nil
}
