package google

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type InstalledBody struct {
	ClientID                string   `json:"client_id"`
	ProjectID               string   `json:"project_id"`
	AuthURI                 string   `json:"auth_uri"`
	TokenURI                string   `json:"token_uri"`
	AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `json:"client_secret"`
	RedirectURIs            []string `json:"redirect_uris"`
}

type Credentials struct {
	Installed InstalledBody `json:"installed"`
}

var gmailSrv *gmail.Service
var driveSrv *drive.Service

func CredentialsFromEnv() ([]byte, error) {
	clientID, err := util.LoadEnvVar("CLIENT_ID")
	if err != nil {
		return nil, err
	}
	clientSecret, err := util.LoadEnvVar("CLIENT_SECRET")
	if err != nil {
		return nil, err
	}
	redirectURI, err := util.LoadEnvVar("REDIRECT_URI")
	if err != nil {
		return nil, err
	}

	credentials := Credentials{
		Installed: InstalledBody{
			ClientID:                clientID,
			ProjectID:               "film-sync",
			AuthURI:                 "https://accounts.google.com/o/oauth2/auth",
			TokenURI:                "https://oauth2.googleapis.com/token",
			AuthProviderX509CertURL: "https://www.googleapis.com/oauth2/v1/certs",
			ClientSecret:            clientSecret,
			RedirectURIs:            []string{redirectURI},
		},
	}

	jsonData, err := json.Marshal(credentials)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials: %v", err)
	}

	return jsonData, nil
}

func Config() (*oauth2.Config, error) {
	b, err := CredentialsFromEnv()
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, drive.DriveFileScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	return config, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	tok, err := database.GetToken()
	if err != nil {
		return nil, fmt.Errorf("no oauth token in db: %v", err)
	}

	return config.Client(ctx, tok), nil
}

func gmailService(ctx context.Context, client *http.Client) error {
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	gmailSrv = srv
	log.Default().Println("[Google] Successfully retrieved Gmail service")
	return nil
}

func driveService(ctx context.Context, client *http.Client) error {
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Google Drive client: %v", err)
	}

	driveSrv = srv
	log.Default().Println("[Google] Successfully retrieved Drive service")
	return nil
}

func StartServices(ctx context.Context) error {
	config, err := Config()
	if err != nil {
		return err
	}

	client, err := getClient(ctx, config)
	if err != nil {
		return err
	}

	err = gmailService(ctx, client)
	if err != nil {
		return err
	}

	err = driveService(ctx, client)
	if err != nil {
		return err
	}

	return nil
}
