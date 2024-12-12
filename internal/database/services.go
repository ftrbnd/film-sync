package database

import (
	"context"
	"fmt"
	"log"
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

func GetScans() ([]FilmScan, error) {
	cur, err := scanCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("unable to get all film scans: %v", err)
	}

	defer cur.Close(context.Background())

	var results []FilmScan
	err = cur.All(context.Background(), &results)
	if err != nil {
		return nil, fmt.Errorf("unable to decode results: %v", err)
	}

	return results, nil
}

func GetOneScan(scanID string) (*FilmScan, error) {
	objectID, err := bson.ObjectIDFromHex(scanID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	res := scanCollection.FindOne(context.Background(), filter)
	if res.Err() != nil {
		return nil, fmt.Errorf(`unable to get film scan "%s": %v`, scanID, res.Err())
	}

	scan := &FilmScan{}
	err = res.Decode(scan)
	if err != nil {
		return nil, err
	}

	return scan, nil
}

func AddScan(f FilmScan) (*mongo.InsertOneResult, error) {
	res, err := scanCollection.InsertOne(context.TODO(), f)
	if err != nil {
		return nil, fmt.Errorf("unable to insert document: %v", err)
	}

	log.Default().Printf("[MongoDB] Inserted new film scan from email: %s", f.EmailID)
	return res, nil
}

func UpdateCldFolderName(old string, new string) error {
	filter := bson.M{"cld_folder_name": old}
	update := bson.M{"$set": bson.M{"cld_folder_name": new}}
	res := scanCollection.FindOneAndUpdate(context.Background(), filter, update)
	if res.Err() != nil {
		return fmt.Errorf(`unable to update film scan with cld_folder_name "%s": %v`, old, res.Err())
	}

	return nil
}

func EmailExists(savedScans []FilmScan, fetchedEmail *gmail.Message) bool {
	exists := slices.ContainsFunc(savedScans, func(saved FilmScan) bool {
		return saved.EmailID == fetchedEmail.Id
	})

	return exists
}

func SaveToken(tok *oauth2.Token) (*mongo.InsertOneResult, error) {
	_, err := tokenCollection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("unable to reset oauth_token collection: %v", err)
	}

	res, err := tokenCollection.InsertOne(context.TODO(), tok)
	if err != nil {
		return nil, fmt.Errorf("unable to save token: %v", err)
	}

	log.Default().Println("[MongoDB] Saved oauth token to database")
	return res, nil
}

func GetToken() (*oauth2.Token, error) {
	res := tokenCollection.FindOne(context.TODO(), bson.D{})

	tok := &oauth2.Token{}
	err := res.Decode(tok)
	if err != nil {
		return nil, err
	}

	return tok, nil
}

func TokenCount() (int64, error) {
	count, err := tokenCollection.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		return 0, err
	}

	return count, nil
}
