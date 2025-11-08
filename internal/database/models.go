package database

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileDocument struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	MsgID          int                `bson:"msg_id" json:"msg_id"`
	ChatID         string             `bson:"chat_id" json:"chat_id"`
	FileID         string             `bson:"file_id" json:"file_id"`
	FileUniqueID   string             `bson:"file_unique_id" json:"file_unique_id"`
	Hash           string             `bson:"hash" json:"hash"`
	MessageURL     string             `bson:"message_url" json:"message_url"`
	RetrievalLink  string             `bson:"retrieval_link" json:"retrieval_link"`
	SessionID      string             `bson:"session_id" json:"session_id"`
	Size           int64              `bson:"size" json:"size"`
	Title          string             `bson:"title" json:"title"`
	Type           string             `bson:"type" json:"type"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}
