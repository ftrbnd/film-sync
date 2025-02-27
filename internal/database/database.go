package database

import (
	"context"
	"log"
	"strings"

	"github.com/ftrbnd/film-sync/internal/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var db *mongo.Database
var scanCollection *mongo.Collection
var tokenCollection *mongo.Collection

func Connect() error {
	uri, err := util.LoadEnvVar("MONGODB_URI")
	if err != nil {
		return err
	}
	dbName := strings.Split(uri, "mongodb.net/")
	dbName = strings.Split(dbName[1], "?")

	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(opts)
	if err != nil {
		return err
	}

	// Send a ping to confirm a successful connection
	db = client.Database(dbName[0])
	err = db.RunCommand(context.Background(), bson.D{{Key: "ping", Value: 1}}).Err()
	if err != nil {
		return err
	}

	scanCollection = db.Collection("scans")
	tokenCollection = db.Collection("oauth_tokens")

	log.Default().Printf("[MongoDB] Successfully connected to %s", dbName[0])
	return nil
}

func Disconnect() error {
	return db.Client().Disconnect(context.Background())
}
