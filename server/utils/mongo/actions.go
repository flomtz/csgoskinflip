package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateOne(database, collectionName string, document any, filter bson.M) error {
	collection := mongoCli.Database(database).Collection(collectionName)

	_, err := collection.UpdateOne(context.Background(), filter, document, &options.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update: %s", err)
	}

	return nil
}

func InsertOne(database, collectionName string, document any) error {
	collection := mongoCli.Database(database).Collection(collectionName)

	_, err := collection.InsertOne(context.Background(), document, &options.InsertOneOptions{})
	if err != nil {
		return fmt.Errorf("failed to insert: %s", err)
	}

	return nil
}

func FindOne[T any](database, collectionName string, filter bson.M) (*T, error) {

	collection := mongoCli.Database(database).Collection(collectionName)

	var data bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("no entry found under %s in collection %s: %v", database, collectionName, err)
	}

	raw, err := bson.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal environment data: %v", err)
	}

	var jsonData T
	err = bson.Unmarshal(raw, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal environment data: %v", err)
	}
	return &jsonData, nil
}

func FindAll(database, collectionName string, filter bson.M) ([]bson.M, error) {
	collection := mongoCli.Database(database).Collection(collectionName)

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries in %s.%s: %v", database, collectionName, err)
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return nil, fmt.Errorf("failed to decode results from %s.%s: %v", database, collectionName, err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no entries found in %s.%s", database, collectionName)
	}

	return results, nil
}

func CheckForExisting(database, collectionName string, checkFor bson.M) (bool, error) {
	collection := mongoCli.Database(database).Collection(collectionName)

	result := collection.FindOne(context.Background(), checkFor)
	err := result.Err()
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			return false, nil
		} else {
			return false, fmt.Errorf("failed to check for existing entry: %v", err)
		}
	}
	return true, nil
}

func DeleteOne(database, collectionName string, filter bson.M) error {
	collection := mongoCli.Database(database).Collection(collectionName)

	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete: %s", err)
	}

	return nil
}
