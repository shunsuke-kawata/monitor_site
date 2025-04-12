package model

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MyMongoDB struct {
	Client *mongo.Client
}

func (db *MyMongoDB) getUri() (uri string) {
	username := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
	password := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
	serviceName := os.Getenv("MONGODB_SERVICE")
	port := os.Getenv("MONGODB_PORT")
	uri = fmt.Sprintf("mongodb://%s:%s@%s:%s", username, password, serviceName, port)
	return uri
}

func (db *MyMongoDB) Connect(ctx context.Context) (err error) {
	opt := options.Client().ApplyURI(db.getUri())
	if err := opt.Validate(); err != nil {
		return err
	}

	db.Client, err = mongo.Connect(ctx, opt)
	return err

}

func (db *MyMongoDB) Disconnect(ctx context.Context) error {
	return db.Client.Disconnect(ctx)
}

func (db *MyMongoDB) Ping(ctx context.Context) error {
	if err := db.Client.Ping(ctx, nil); err != nil {
		return err
	}
	fmt.Println("Successfully connect")
	return nil
}

func (db *MyMongoDB) GetOrCreateCollection(ctx context.Context, databaseName, collectionName string) (*mongo.Collection, error) {
	database := db.Client.Database(databaseName)

	// 既存のコレクション一覧を取得
	collections, err := database.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	// コレクションが存在するかチェック
	exists := false
	for _, name := range collections {
		if name == collectionName {
			exists = true
			break
		}
	}

	// 存在しなければ作成
	if !exists {
		if err := database.CreateCollection(ctx, collectionName); err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
		fmt.Printf("Collection '%s' created in database '%s'\n", collectionName, databaseName)
	}

	// コレクションを返す
	return database.Collection(collectionName), nil
}

func (db *MyMongoDB) ListCollections(ctx context.Context, databaseName string) ([]string, error) {
	database := db.Client.Database(databaseName)
	collections, err := database.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	return collections, nil
}

func (db *MyMongoDB) DropCollection(ctx context.Context, databaseName, collectionName string) error {
	database := db.Client.Database(databaseName)
	if err := database.Collection(collectionName).Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop collection '%s': %w", collectionName, err)
	}
	fmt.Printf("Collection '%s' dropped from database '%s'\n", collectionName, databaseName)
	return nil
}

func (db *MyMongoDB) FindAllDocuments(ctx context.Context, databaseName, collectionName string) ([]map[string]interface{}, error) {
	collection := db.Client.Database(databaseName).Collection(collectionName)

	cursor, err := collection.Find(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to find documents in collection '%s': %w", collectionName, err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		results = append(results, doc)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return results, nil
}
