package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Indexing MongoDB collections for better performance
func CreateProductIndexes(client *mongo.Client) {
	productCol := ProductData(client, "Products")
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "product_name", Value: "text"},
			{Key: "category", Value: 1},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := productCol.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		fmt.Println("Warning: Could not create index:", err)
	} else {
		fmt.Println("Successfully optimized Product indexes")
	}
}

func DBSet() *mongo.Client {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	client, err := mongo.NewClient(options.Client().ApplyURI(dbURL))
	fmt.Println("DB URL:", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err, "Failed to connect to mongoDB")
		return nil
	}

	fmt.Println("Successfully connected to mongoDB")

	CreateProductIndexes(client)
	return client
}

var Client *mongo.Client = DBSet()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return collection
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {
	var productCollection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return productCollection
}
