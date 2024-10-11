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

func GetEmails() ([]Email, error) {
	collection := GetCollection("emails")

	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("unable to get all emails: %v", err)
	}

	defer cur.Close(context.Background())

	var results []Email
	err = cur.All(context.Background(), &results)
	if err != nil {
		return nil, fmt.Errorf("unable to decode results: %v", err)
	}

	return results, nil
}

func AddEmail(e Email) (*mongo.InsertOneResult, error) {
	collection := GetCollection("emails")

	res, err := collection.InsertOne(context.TODO(), e)
	if err != nil {
		return nil, fmt.Errorf("unable to insert document: %v", err)
	}

	log.Default().Printf("[MongoDB] Inserted email %s", e.EmailID)
	return res, nil
}

func EmailExists(savedEmails []Email, fetchedEmail *gmail.Message) bool {
	exists := slices.ContainsFunc(savedEmails, func(saved Email) bool {
		return saved.EmailID == fetchedEmail.Id
	})

	return exists
}

func SaveToken(tok *oauth2.Token) (*mongo.InsertOneResult, error) {
	collection := GetCollection("oauth_tokens")

	_, err := collection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("unable to reset oauth_token collection: %v", err)
	}

	res, err := collection.InsertOne(context.TODO(), tok)
	if err != nil {
		return nil, fmt.Errorf("unable to save token: %v", err)
	}

	log.Default().Println("[MongoDB] Saved oauth token to database")
	return res, nil
}

func GetToken() (*oauth2.Token, error) {
	collection := GetCollection("oauth_tokens")

	res := collection.FindOne(context.TODO(), bson.D{})

	tok := &oauth2.Token{}
	err := res.Decode(tok)
	if err != nil {
		return nil, err
	}

	return tok, nil
}

func TokenCount() (int64, error) {
	collection := GetCollection("oauth_tokens")

	count, err := collection.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		return 0, err
	}

	return count, nil
}
