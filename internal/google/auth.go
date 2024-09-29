package google

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, acr chan *oauth2.Token) (*http.Client, error) {
	tok, err := database.GetToken()
	if err != nil {
		err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		tok = <-acr
	}

	return config.Client(context.Background(), tok), nil
}

func getTokenFromWeb(config *oauth2.Config) error {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	err := discord.SendAuthMessage(authURL)
	if err != nil {
		return err
	}

	log.Default().Println("Waiting for user to authenticate...")
	return nil
}

func CredentialsFromEnv() ([]byte, error) {
	clientID, err := util.LoadEnvVar("CLIENT_ID")
	if err != nil {
		return nil, err
	}
	projectID, err := util.LoadEnvVar("PROJECT_ID")
	if err != nil {
		return nil, err
	}
	authURI, err := util.LoadEnvVar("AUTH_URI")
	if err != nil {
		return nil, err
	}
	tokenURI, err := util.LoadEnvVar("TOKEN_URI")
	if err != nil {
		return nil, err
	}
	authProviderX509CertURL, err := util.LoadEnvVar("AUTH_PROVIDER_X509_CERT_URL")
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
			ProjectID:               projectID,
			AuthURI:                 authURI,
			TokenURI:                tokenURI,
			AuthProviderX509CertURL: authProviderX509CertURL,
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

func GmailService(acr chan *oauth2.Token) error {
	ctx := context.Background()

	config, err := Config()
	if err != nil {
		return err
	}
	client, err := getClient(config, acr)
	if err != nil {
		return err
	}

	gmailSrv, err = gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	log.Default().Println("[Gmail] Successfully retrieved service")
	return nil
}

func DriveService(acr chan *oauth2.Token) error {
	ctx := context.Background()

	config, err := Config()
	if err != nil {
		return err
	}
	client, err := getClient(config, acr)
	if err != nil {
		return err
	}

	driveSrv, err = drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Google Drive client: %v", err)
	}

	log.Default().Println("[Google Drive] Successfully retrieved service")
	return nil
}
