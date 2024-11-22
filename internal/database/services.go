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
	filter := bson.M{"folder_name": bson.M{"$exists": true}}
	cur, err := scanCollection.Find(context.Background(), filter)
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

func GetOneScan(folder string) (*FilmScan, error) {
	filter := bson.M{"folder_name": folder}
	res := scanCollection.FindOne(context.Background(), filter)
	if res.Err() != nil {
		return nil, fmt.Errorf(`unable to get film scan "%s": %v`, folder, res.Err())
	}

	scan := &FilmScan{}
	err := res.Decode(scan)
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

func UpdateFolderName(old string, new string) (*mongo.UpdateResult, error) {
	filter := bson.M{"folder_name": old}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "folder_name", Value: new},
	}}}

	res, err := scanCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, err
	}

	log.Default().Printf("[MongoDB] Updated folder name to %s", new)
	return res, nil
}

func AddImageKeysToScan(downloadLink string, folder string, keys []string) (*mongo.UpdateResult, error) {
	filter := bson.M{"download_link": downloadLink}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "image_keys", Value: keys},
		{Key: "folder_name", Value: folder},
	}}}

	res, err := scanCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, err
	}

	log.Default().Printf("[MongoDB] Saved %d image keys to document", len(keys))
	return res, nil
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
