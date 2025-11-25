package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoCli *mongo.Client

func InitMongoClient(connectionURL string) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectionURL))
	if err != nil {
		panic(err)
	}
	mongoCli = client
}
