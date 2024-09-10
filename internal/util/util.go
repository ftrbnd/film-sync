package util

import (
	"log"
	"os"
)

func LoadEnvVar(key string) string {
	v, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("%s environment variable not found", key)
	}

	return v
}