package server

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/files"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
)

func startJob(links []string) error {
	dst := "output"
	format := "tif"

	for _, link := range links {
		z, err := files.DownloadFrom(link)
		if err != nil {
			return fmt.Errorf("failed to download from link: %v", err)
		}

		files.Unzip(z, dst, format)
		c, err := files.ConvertToPNG(format, dst)
		if err != nil {
			return fmt.Errorf("failed to convert to png: %v", err)
		}

		s3Folder, driveFolderID, message, err := files.Upload(dst, z, c)
		if err != nil {
			return fmt.Errorf("failed to upload files: %v", err)
		}

		err = discord.SendSuccessMessage(s3Folder, driveFolderID, message)
		if err != nil {
			return fmt.Errorf("failed to send discord success message: %v", err)
		}
	}

	log.Default().Println("[HTTP] Finished running daily job!")
	return nil
}

func verify(body []byte, tokenString, signingKey string) error {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(signingKey), nil
		})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("invalid token")
	}

	if !claims.VerifyIssuer("Upstash", true) {
		return fmt.Errorf("invalid issuer")
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return fmt.Errorf("token has expired")
	}
	if !claims.VerifyNotBefore(time.Now().Unix(), true) {
		return fmt.Errorf("token is not valid yet")
	}

	bodyHash := sha256.Sum256(body)
	if claims["body"] != base64.URLEncoding.EncodeToString(bodyHash[:]) {
		return fmt.Errorf("body hash does not match")
	}

	return nil
}

func dailyHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("[HTTP] Received /daily request")

	env, _ := util.LoadEnvVar("GO_ENV")
	if env != "development" {
		currentSigningKey, err := util.LoadEnvVar("QSTASH_CURRENT_SIGNING_KEY")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		nextSigningKey, err := util.LoadEnvVar("QSTASH_NEXT_SIGNING_KEY")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tokenString := r.Header.Get("Upstash-Signature")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = verify(body, tokenString, currentSigningKey)
		if err != nil {
			log.Default().Printf("[HTTP] Unable to verify signature with current signing key: %v", err)
			err = verify(body, tokenString, nextSigningKey)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Request accepted for processing")

	go func() {
		newLinks, err := google.CheckEmail()
		if err != nil {
			discord.SendErrorMessage(err)

			authErr := strings.Contains(err.Error(), "service hasn't been initialized") || strings.Contains(err.Error(), "token expired")
			if authErr {
				config, _ := google.Config()
				authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
				discord.SendAuthMessage(authURL)
				log.Default().Println("[Google] Sent auth request to user via Discord")
			}

			return
		}

		log.Default().Printf("[HTTP] Found %d new links", len(newLinks))

		if len(newLinks) > 0 {
			err = startJob(newLinks)
			if err != nil {
				discord.SendErrorMessage(err)
			}
		}
	}()
}
