package http

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello film-sync!")
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("[HTTP] Received /auth request")

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "Missing code or state", http.StatusUnauthorized)
		return
	}

	tok, err := googleConfig.Exchange(ctx, code, oauth2.AccessTypeOffline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	database.SaveToken(tok)
	google.StartServices(ctx, googleConfig)

	fmt.Fprintln(w, "Thank you! You can now close this tab.")
}

func dailyHandler(w http.ResponseWriter, r *http.Request, runDailyJob func() error) {
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

	go runDailyJob()
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
