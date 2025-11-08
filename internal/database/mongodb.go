package database

import (
	"EverythingSuckz/fsb/internal/utils"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	
	log.Info("=== MONGODB INITIALIZATION STARTED ===")
	log.Info("MongoDB Configuration", 
		zap.String("URI", mongoURI), 
		zap.String("Database", dbName), 
		zap.String("Collection", collectionName))
	
	if mongoURI == "" {
		log.Warn("⚠️  MongoDB URI not provided, skipping MongoDB initialization")
		log.Warn("   To enable MongoDB, set MONGO_URI in your fsb.env file")
		return nil
	}

	log.Info("Connecting to MongoDB...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Error("❌ Failed to connect to MongoDB", zap.Error(err))
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	log.Info("✓ MongoDB client connected")

	// Ping the database to verify connection
	log.Info("Pinging MongoDB to verify connection...")
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Error("❌ Failed to ping MongoDB", zap.Error(err))
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	log.Info("✓ MongoDB ping successful")

	Client = client
	Database = client.Database(dbName)
	Collection = Database.Collection(collectionName)

	log.Info("✅ Successfully connected to MongoDB!", 
		zap.String("database", dbName), 
		zap.String("collection", collectionName))

	// Create indexes
	log.Info("Creating database indexes...")
	if err := createIndexes(ctx, log); err != nil {
		log.Error("❌ Failed to create indexes", zap.Error(err))
	}

	log.Info("=== MONGODB INITIALIZATION COMPLETED ===")
	return nil
}

func createIndexes(ctx context.Context, log *zap.Logger) error {
	log.Info("Creating indexes for fields: msg_id, hash, file_id, session_id")
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

	indexNames, err := Collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Error("Failed to create indexes", zap.Error(err))
		return err
	}

	log.Info("✓ MongoDB indexes created successfully", zap.Int("count", len(indexNames)))
	return nil
}

func SaveFile(ctx context.Context, doc *FileDocument) error {
	if Collection == nil {
		return fmt.Errorf("MongoDB not initialized")
	}

	doc.CreatedAt = time.Now()
	result, err := Collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	
	// Log success with details
	log := utils.Logger.Named("MongoDB")
	log.Info("💾 File saved to MongoDB",
		zap.String("_id", result.InsertedID.(primitive.ObjectID).Hex()),
		zap.Int("msg_id", doc.MsgID),
		zap.String("title", doc.Title),
		zap.Int64("size", doc.Size),
		zap.String("type", doc.Type),
		zap.String("hash", doc.Hash),
		zap.String("session_id", doc.SessionID))
	
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
