package utils

import (
	"encoding/base64"
	"fmt"

	"github.com/gotd/td/tg"
)

// GetTelegramFileID extracts the Telegram file_id from document
func GetTelegramFileID(doc *tg.Document) string {
	if doc == nil {
		return ""
	}
	// Create a simple representation of the file ID
	// In production, you might want to use the actual Telegram file_id format
	return fmt.Sprintf("BAACAgQAAyEFAA%s", encodeFileID(doc.ID, doc.AccessHash))
}

// GetTelegramPhotoFileID extracts the Telegram file_id from photo
func GetTelegramPhotoFileID(photo *tg.Photo) string {
	if photo == nil {
		return ""
	}
	return fmt.Sprintf("BAACAgQAAyEFAA%s", encodeFileID(photo.ID, photo.AccessHash))
}

// GetFileUniqueID creates a unique identifier for the file
func GetFileUniqueID(fileReference []byte) string {
	if len(fileReference) > 12 {
		return base64.RawURLEncoding.EncodeToString(fileReference[:12])
	}
	return base64.RawURLEncoding.EncodeToString(fileReference)
}

func encodeFileID(id int64, accessHash int64) string {
	combined := fmt.Sprintf("%d_%d", id, accessHash)
	return base64.RawURLEncoding.EncodeToString([]byte(combined))[:16]
}
