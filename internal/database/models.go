package database

import "go.mongodb.org/mongo-driver/v2/bson"

type FilmScan struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	EmailID      string        `bson:"email_id,omitempty"`
	DownloadLink string        `bson:"download_link,omitempty"`
	FolderName   string        `bson:"folder_name,omitempty"`
	ImageKeys    []string      `bson:"image_keys,omitempty"`
}
