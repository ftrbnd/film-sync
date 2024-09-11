package gmail

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/api/gmail/v1"
)

func getEmailsBySender(sender string, service *gmail.Service) []*gmail.Message {
	q := fmt.Sprintf("from:%s", sender)

	res, err := service.Users.Messages.List("me").Q(q).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages from %s: %v", sender, err)
	}

	return res.Messages
}

func filterEmailsByMetadata(messages []*gmail.Message, fieldName string, fieldValue string, service *gmail.Service) []*gmail.Message {
	var emails []*gmail.Message

	for _, msg := range messages {
		message, err := service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message_%s: %v", msg.Id, err)
		}

		for _, header := range message.Payload.Headers {
			if header.Name == fieldName {
				if header.Value == fieldValue {
					emails = append(emails, msg)
				}
			}
		}
	}

	return emails
}

func GetDownloadLink(message *gmail.Message, service *gmail.Service) string {
	msg, err := service.Users.Messages.Get("me", message.Id).Format("full").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve message: %v", err)
	}

	data := msg.Payload.Parts[0].Body.Data
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		log.Fatalf("Unable to decode message body: %v", err)
	}

	lines := strings.Split(string(decoded), "\n")
	link := lines[6] // or find by index of https://wetransfter.com/downloads

	return link
}

func FetchEmails(s *gmail.Service) []*gmail.Message {
	fromEmail := util.LoadEnvVar("FROM_EMAIL")
	replyToEmail := util.LoadEnvVar("REPLY_TO_EMAIL")

	e := getEmailsBySender(fromEmail, s)
	f := filterEmailsByMetadata(e, "Reply-To", replyToEmail, s)

	return f
}

func CheckEmail(c *mongo.Client, s *gmail.Service) []string {
	log.Default().Println("Checking email...")

	emails := FetchEmails(s)
	saved := database.GetEmails(c)

	var newLinks []string

	for _, email := range emails {
		exists := database.EmailExists(saved, email)
		if !exists {
			link := GetDownloadLink(email, s)

			newEmail := database.Email{ID: bson.NewObjectID(), EmailID: email.Id, DownloadLink: link}

			database.AddEmail(c, newEmail)
			log.Default().Printf("Added email #%s", email.Id)

			newLinks = append(newLinks, link)
		} else {
			log.Default().Printf("Email #%s already exists in database", email.Id)
		}
	}

	return newLinks
}
