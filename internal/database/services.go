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

func GetEmails(c *mongo.Client) ([]Email, error) {
	collection := EmailCollection(c)

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

func AddEmail(c *mongo.Client, e Email) (*mongo.InsertOneResult, error) {
	collection := EmailCollection(c)

	res, err := collection.InsertOne(context.TODO(), e)
	if err != nil {
		return nil, fmt.Errorf("unable to insert document: %v", err)
	}

	log.Default().Printf("Inserted email %s to database", e.EmailID)
	return res, nil
}

func EmailExists(savedEmails []Email, fetchedEmail *gmail.Message) bool {
	exists := slices.ContainsFunc(savedEmails, func(saved Email) bool {
		return saved.EmailID == fetchedEmail.Id
	})

	return exists
}

func SaveToken(c *mongo.Client, tok *oauth2.Token) (*mongo.InsertOneResult, error) {
	collection := OAuthTokenCollection(c)

	_, err := collection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("unable to reset oauth_token collection: %v", err)
	}

	res, err := collection.InsertOne(context.TODO(), tok)
	if err != nil {
		return nil, fmt.Errorf("unable to save token: %v", err)
	}

	log.Default().Println("Saved token to database")
	return res, nil
}

func GetToken(c *mongo.Client) (*oauth2.Token, error) {
	collection := OAuthTokenCollection(c)

	res := collection.FindOne(context.TODO(), bson.D{})

	tok := &oauth2.Token{}
	err := res.Decode(tok)

	return tok, err
}
