package google

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, acr chan *oauth2.Token) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := getTokenFromFile("token.json")
	if err != nil {
		getTokenFromWeb(config)
		tok = <-acr
	}

	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	content := fmt.Sprintf("Go to the following link in your browser: \n%v", authURL)
	discord.SendMessage(content)

	log.Default().Println("Waiting for user to authenticate...")
}

// Retrieves a token from a local file.
func getTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func SaveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	util.CheckError("Unable to cache oauth token", err)

	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func CredentialsFromEnv() []byte {
	credentials := Credentials{
		Installed: InstalledBody{
			ClientID:                util.LoadEnvVar("CLIENT_ID"),
			ProjectID:               util.LoadEnvVar("PROJECT_ID"),
			AuthURI:                 util.LoadEnvVar("AUTH_URI"),
			TokenURI:                util.LoadEnvVar("TOKEN_URI"),
			AuthProviderX509CertURL: util.LoadEnvVar("AUTH_PROVIDER_X509_CERT_URL"),
			ClientSecret:            util.LoadEnvVar("CLIENT_SECRET"),
			RedirectURIs:            []string{util.LoadEnvVar("REDIRECT_URI")},
		},
	}

	jsonData, err := json.Marshal(credentials)
	util.CheckError("Unable to read credentials", err)

	return jsonData
}

func Config() *oauth2.Config {
	b := CredentialsFromEnv()

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, drive.DriveFileScope)
	util.CheckError("Unable to parse client secret file to config", err)

	return config
}

func GmailService(acr chan *oauth2.Token) *gmail.Service {
	ctx := context.Background()

	config := Config()
	client := getClient(config, acr)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	util.CheckError("Unable to retrieve Gmail client", err)

	log.Default().Println("Successfully retrieved Gmail service!")
	return service
}

func DriveService(acr chan *oauth2.Token) *drive.Service {
	ctx := context.Background()

	config := Config()
	client := getClient(config, acr)

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	util.CheckError("Unable to retrieve Google Drive client", err)

	log.Default().Println("Successfully retrieved Google Drive service!")
	return service
}
