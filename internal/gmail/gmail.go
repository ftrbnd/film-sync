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