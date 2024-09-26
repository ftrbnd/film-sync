package google

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/ftrbnd/film-sync/internal/database"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/util"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
func getClient(config *oauth2.Config, acr chan *oauth2.Token, client *mongo.Client, bot *discordgo.Session) *http.Client {
	tok, err := getTokenFromDatabase(client)
	if err != nil {
		getTokenFromWeb(config, bot)
		tok = <-acr
	}

	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config, bot *discordgo.Session) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	discord.SendAuthMessage(authURL, bot)

	log.Default().Println("Waiting for user to authenticate...")
}

func getTokenFromDatabase(client *mongo.Client) (*oauth2.Token, error) {
	tok, err := database.GetToken(client)
	return tok, err
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

func GmailService(acr chan *oauth2.Token, db *mongo.Client, bot *discordgo.Session) *gmail.Service {
	ctx := context.Background()

	config := Config()
	client := getClient(config, acr, db, bot)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	util.CheckError("Unable to retrieve Gmail client", err)

	log.Default().Println("[Gmail] Successfully retrieved service")
	return service
}

func DriveService(acr chan *oauth2.Token, db *mongo.Client, bot *discordgo.Session) *drive.Service {
	ctx := context.Background()

	config := Config()
	client := getClient(config, acr, db, bot)

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	util.CheckError("Unable to retrieve Google Drive client", err)

	log.Default().Println("[Google Drive] Successfully retrieved service")
	return service
}
