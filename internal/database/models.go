package database

import "go.mongodb.org/mongo-driver/v2/bson"

type FilmScan struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	EmailID       string        `bson:"email_id,omitempty" json:"email_id,omitempty"`
	DownloadURL   string        `bson:"download_url,omitempty" json:"download_url,omitempty"`
	DriveFolderID string        `bson:"drive_folder_id,omitempty" json:"drive_folder_id,omitempty"`
	CldFolderName string        `bson:"cld_folder_name,omitempty" json:"cld_folder_name,omitempty"`
}
