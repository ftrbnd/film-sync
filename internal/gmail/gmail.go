package gmail

import (
	"fmt"
	"log"
	"os"
	"time"
)

func loadClientID() string {
  clientID, exists := os.LookupEnv("GMAIL_CLIENT_ID")
	if !exists {
		log.Fatal("GMAIL_CLIENT_ID environment variable not found")
	}

	return clientID
}

func checkEmail(t time.Time) {
	fmt.Println(t, "Checking email...")
	service := GetGmailService()
	
	log.Default().Println("Gmail service:", service)

	user := "me"

	res, err := service.Users.Messages.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}

	log.Default().Printf("Found %d messages", len(res.Messages))
	
}

func ScheduleJob() {
	clientID := loadClientID()
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