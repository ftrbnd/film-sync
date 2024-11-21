package database

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Email struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	EmailID      string        `bson:"emailID,omitempty"`
	DownloadLink string        `bson:"downloadLink,omitempty"`
	ImageKeys    []string      `bson:"imageKeys,omitempty"`
}
