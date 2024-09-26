package database

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Email struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	EmailID      string        `bson:"emailID,omitempty"`
	DownloadLink string        `bson:"downloadLink,omitempty"`
}

func EmailCollection(c *mongo.Client) *mongo.Collection {
	collection := c.Database("film-sync").Collection("emails")
	return collection
}

func OAuthTokenCollection(c *mongo.Client) *mongo.Collection {
	collection := c.Database("film-sync").Collection("oauth_tokens")
	return collection
}
