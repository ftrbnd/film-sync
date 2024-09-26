package database

import (
	"context"
	"log"
	"slices"

	"github.com/ftrbnd/film-sync/internal/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

func GetEmails(c *mongo.Client) []Email {
	collection := EmailCollection(c)

	cur, err := collection.Find(context.Background(), bson.D{})
	util.CheckError("Unable to get all emails", err)

	defer cur.Close(context.Background())

	var results []Email
	err = cur.All(context.Background(), &results)
	util.CheckError("Unable to decode results", err)

	return results
}

func AddEmail(c *mongo.Client, e Email) *mongo.InsertOneResult {
	collection := EmailCollection(c)

	res, err := collection.InsertOne(context.TODO(), e)
	util.CheckError("Unable to insert document", err)

	log.Default().Printf("Inserted email %s to database", e.EmailID)
	return res
}

func EmailExists(savedEmails []Email, fetchedEmail *gmail.Message) bool {
	exists := slices.ContainsFunc(savedEmails, func(saved Email) bool {
		return saved.EmailID == fetchedEmail.Id
	})

	return exists
}

func SaveToken(c *mongo.Client, tok *oauth2.Token) *mongo.InsertOneResult {
	collection := OAuthTokenCollection(c)

	_, err := collection.DeleteMany(context.TODO(), bson.D{})
	util.CheckError("Unable to reset OAuthToken collection", err)

	res, err := collection.InsertOne(context.TODO(), tok)
	util.CheckError("Unable to save token", err)

	log.Default().Println("Saved token to database")
	return res
}

func GetToken(c *mongo.Client) (*oauth2.Token, error) {
	collection := OAuthTokenCollection(c)

	res := collection.FindOne(context.TODO(), bson.D{})

	tok := &oauth2.Token{}
	err := res.Decode(tok)

	return tok, err
}
