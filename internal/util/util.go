package util

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	prod := os.Getenv("PROD")

	if prod != "true" {
		err := godotenv.Load()
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadEnvVar(key string) (string, error) {
	v, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("%s environment variable not found", key)
	}

	return v, nil
}

func CheckError(m string, e error) {
	if e != nil {
		log.Fatalf("%s: %v", m, e)
	}
}
