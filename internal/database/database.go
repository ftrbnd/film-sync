package database

import (
	"context"
	"log"

	"github.com/ftrbnd/film-sync/internal/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect() (*mongo.Client, error) {
	uri, err := util.LoadEnvVar("MONGODB_URI")
	if err != nil {
		return nil, err
	}

	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	// Send a ping to confirm a successful connection
	err = client.Database("film-sync").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err()
	if err != nil {
		return nil, err
	}

	log.Default().Println("[MongoDB] Successfully connected")
	return client, nil
}
