package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
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

	database.SaveToken(ctx, tok)
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

func notifyHandler(w http.ResponseWriter, r *http.Request, notify func(scanID, emailID string) error) {
	log.Default().Println("[HTTP] Received /notify request")

	env, _ := util.LoadEnvVar("GO_ENV")
	if env != "development" {
		http.Error(w, "notify is only available in development", http.StatusForbidden)
		return
	}

	if notify == nil {
		http.Error(w, "notify handler not configured", http.StatusInternalServerError)
		return
	}

	scanID := r.URL.Query().Get("scan_id")
	emailID := r.URL.Query().Get("email_id")
	if scanID == "" && emailID == "" {
		http.Error(w, "provide scan_id or email_id query param", http.StatusBadRequest)
		return
	}

	if err := notify(scanID, emailID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Success notification resent")
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

func SendDeployRequest(message string) error {
	buildHookURL, err := util.LoadEnvVar("NETLIFY_BUILD_HOOK_URL")
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?trigger_title=%s&clear_cache=true", buildHookURL, url.QueryEscape(message))
	resp, err := http.Post(url, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("failed to send deploy request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("did not receive 200 OK status")
	}

	return nil
}

// NotifyFilmOrient asks gio-hub to inspect a Cloudinary folder for sideways
// portrait frames and rotate them. Expects 202 Accepted; processing continues
// asynchronously on gio-hub.
func NotifyFilmOrient(folder string) error {
	baseURL, err := util.LoadEnvVar("GIO_HUB_URL")
	if err != nil {
		return err
	}
	secret, err := util.LoadEnvVar("GIO_HUB_API_SECRET")
	if err != nil {
		return err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/film/orient"
	body, err := json.Marshal(map[string]string{"folder": folder})
	if err != nil {
		return fmt.Errorf("failed to marshal film orient payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build film orient request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping gio-hub film orient: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("gio-hub film orient returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	log.Default().Printf("[HTTP] Notified gio-hub to orient folder %q", folder)
	return nil
}
