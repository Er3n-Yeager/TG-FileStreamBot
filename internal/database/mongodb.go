package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
)

type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	log        *zap.Logger
}

func InitMongoDB(log *zap.Logger, mongoURI, dbName, collectionName string) error {
	log = log.Named("MongoDB")
	
	if mongoURI == "" {
		log.Warn("MongoDB URI not provided, skipping MongoDB initialization")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	Client = client
	Database = client.Database(dbName)
	Collection = Database.Collection(collectionName)

	log.Info("Successfully connected to MongoDB", 
		zap.String("database", dbName), 
		zap.String("collection", collectionName))

	// Create indexes
	if err := createIndexes(ctx, log); err != nil {
		log.Error("Failed to create indexes", zap.Error(err))
	}

	return nil
}

func createIndexes(ctx context.Context, log *zap.Logger) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "msg_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: "hash", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: "file_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
	}

	_, err := Collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	log.Info("MongoDB indexes created successfully")
	return nil
}

func SaveFile(ctx context.Context, doc *FileDocument) error {
	if Collection == nil {
		return fmt.Errorf("MongoDB not initialized")
	}

	doc.CreatedAt = time.Now()
	_, err := Collection.InsertOne(ctx, doc)
	return err
}

func GetFileByMsgID(ctx context.Context, msgID int) (*FileDocument, error) {
	if Collection == nil {
		return nil, fmt.Errorf("MongoDB not initialized")
	}

	var doc FileDocument
	err := Collection.FindOne(ctx, bson.M{"msg_id": msgID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func GetFileByHash(ctx context.Context, hash string) (*FileDocument, error) {
	if Collection == nil {
		return nil, fmt.Errorf("MongoDB not initialized")
	}

	var doc FileDocument
	err := Collection.FindOne(ctx, bson.M{"hash": hash}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func UpdateFile(ctx context.Context, msgID int, update bson.M) error {
	if Collection == nil {
		return fmt.Errorf("MongoDB not initialized")
	}

	_, err := Collection.UpdateOne(
		ctx,
		bson.M{"msg_id": msgID},
		bson.M{"$set": update},
	)
	return err
}

func DeleteFile(ctx context.Context, msgID int) error {
	if Collection == nil {
		return fmt.Errorf("MongoDB not initialized")
	}

	_, err := Collection.DeleteOne(ctx, bson.M{"msg_id": msgID})
	return err
}

func GetFilesBySessionID(ctx context.Context, sessionID string) ([]FileDocument, error) {
	if Collection == nil {
		return nil, fmt.Errorf("MongoDB not initialized")
	}

	var files []FileDocument
	cursor, err := Collection.Find(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &files); err != nil {
		return nil, err
	}

	return files, nil
}

func CloseMongoDB(ctx context.Context) error {
	if Client == nil {
		return nil
	}
	return Client.Disconnect(ctx)
}
