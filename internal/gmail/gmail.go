package gmail

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ftrbnd/film-sync/internal/util"
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

func getDownloadLink(message *gmail.Message, service *gmail.Service) string {
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

func checkEmail(t time.Time) {
	fmt.Println(t, "Checking email...")

	service := GetGmailService()
	fromEmail := util.LoadEnvVar("FROM_EMAIL")
	replyToEmail := util.LoadEnvVar("REPLY_TO_EMAIL")
	
	e := getEmailsBySender(fromEmail, service)
	f := filterEmailsByMetadata(e, "Reply-To", replyToEmail, service)

	for _, email := range(f) {
		link := getDownloadLink(email, service)
		log.Default().Println(link)
	}
}

func ScheduleJob() {
	clientID := util.LoadEnvVar("GMAIL_CLIENT_ID")
	log.Default().Println("Client ID: ", clientID)

    ticker := time.NewTicker(5 * time.Second)
    done := make(chan bool)

    go func() {
        for {
            select {
            case <-done:
                return
            case t := <-ticker.C:
				checkEmail(t)
            }
        }
    }()
}