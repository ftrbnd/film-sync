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
		log.Fatal("GMAIL_CLIENT_ID not found")
	}

	return clientID
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
                fmt.Println("Tick at", t)
            }
        }
    }()
}