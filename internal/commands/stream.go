package commands

import (
	"fmt"
	"strconv"
	"strings"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/database"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/celestix/gotgproto/types"
	"github.com/google/uuid"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (m *command) LoadStream(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(
		handlers.NewMessage(nil, sendLink),
	)
}

func supportedMediaFilter(m *types.Message) (bool, error) {
	if not := m.Media == nil; not {
		return false, dispatcher.EndGroups
	}
	switch m.Media.(type) {
	case *tg.MessageMediaDocument:
		return true, nil
	case *tg.MessageMediaPhoto:
		return true, nil
	case tg.MessageMediaClass:
		return false, dispatcher.EndGroups
	default:
		return false, nil
	}
}

func sendLink(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()
	peerChatId := ctx.PeerStorage.GetPeerById(chatId)
	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", nil)
		return dispatcher.EndGroups
	}
	supported, err := supportedMediaFilter(u.EffectiveMessage)
	if err != nil {
		return err
	}
	if !supported {
		ctx.Reply(u, "Sorry, this message type is unsupported.", nil)
		return dispatcher.EndGroups
	}
	update, err := utils.ForwardMessages(ctx, chatId, config.ValueOf.LogChannelID, u.EffectiveMessage.ID)
	if err != nil {
		utils.Logger.Sugar().Error(err)
		ctx.Reply(u, fmt.Sprintf("Error - %s", err.Error()), nil)
		return dispatcher.EndGroups
	}
	messageID := update.Updates[0].(*tg.UpdateMessageID).ID
	doc := update.Updates[1].(*tg.UpdateNewChannelMessage).Message.(*tg.Message).Media
	file, err := utils.FileFromMedia(doc)
	if err != nil {
		ctx.Reply(u, fmt.Sprintf("Error - %s", err.Error()), nil)
		return dispatcher.EndGroups
	}
	fullHash := utils.PackFile(
		file.FileName,
		file.FileSize,
		file.MimeType,
		file.ID,
	)
	hash := utils.GetShortHash(fullHash)
	link := fmt.Sprintf("%s/stream/%d?hash=%s", config.ValueOf.Host, messageID, hash)
	
	// Save to MongoDB
	utils.Logger.Info("📝 Preparing to save file to MongoDB...",
		zap.Int("msg_id", messageID),
		zap.String("filename", file.FileName),
		zap.Int64("size", file.FileSize))
	
	sessionID := uuid.New().String()
	chatIDStr := strconv.FormatInt(config.ValueOf.LogChannelID, 10)
	messageURL := fmt.Sprintf("https://t.me/c/%s/%d", strings.TrimPrefix(chatIDStr, "-100"), messageID)
	retrievalLink := fmt.Sprintf("%s/stream/%d?hash=%s", config.ValueOf.Host, messageID, hash)
	
	// Get file_id and file_unique_id from the media
	var fileID, fileUniqueID string
	switch media := doc.(type) {
	case *tg.MessageMediaDocument:
		if document, ok := media.Document.(*tg.Document); ok {
			fileID = utils.GetTelegramFileID(document)
			fileUniqueID = utils.GetFileUniqueID(document.FileReference)
		}
	case *tg.MessageMediaPhoto:
		if photo, ok := media.Photo.(*tg.Photo); ok {
			fileID = utils.GetTelegramPhotoFileID(photo)
			fileUniqueID = utils.GetFileUniqueID(photo.FileReference)
		}
	}
	
	fileDoc := &database.FileDocument{
		MsgID:         messageID,
		ChatID:        chatIDStr,
		FileID:        fileID,
		FileUniqueID:  fileUniqueID,
		Hash:          hash,
		MessageURL:    messageURL,
		RetrievalLink: retrievalLink,
		SessionID:     sessionID,
		Size:          file.FileSize,
		Title:         file.FileName,
		Type:          file.MimeType,
	}
	
	utils.Logger.Info("💾 Attempting to save to MongoDB...",
		zap.String("collection", "files"),
		zap.String("session_id", sessionID))
	
	if err := database.SaveFile(ctx, fileDoc); err != nil {
		utils.Logger.Error("❌ Failed to save file to MongoDB", zap.Error(err))
	} else {
		utils.Logger.Info("✅ File successfully saved to MongoDB!",
			zap.Int("msg_id", messageID),
			zap.String("session_id", sessionID),
			zap.String("hash", hash))
	}
	
	text := []styling.StyledTextOption{styling.Code(link)}
	row := tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonURL{
				Text: "Download",
				URL:  link + "&d=true",
			},
		},
	}
	if strings.Contains(file.MimeType, "video") || strings.Contains(file.MimeType, "audio") || strings.Contains(file.MimeType, "pdf") {
		row.Buttons = append(row.Buttons, &tg.KeyboardButtonURL{
			Text: "Stream",
			URL:  link,
		})
	}
	markup := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{row},
	}
	if strings.Contains(link, "http://localhost") {
		_, err = ctx.Reply(u, text, &ext.ReplyOpts{
			NoWebpage:        false,
			ReplyToMessageId: u.EffectiveMessage.ID,
		})
	} else {
		_, err = ctx.Reply(u, text, &ext.ReplyOpts{
			Markup:           markup,
			NoWebpage:        false,
			ReplyToMessageId: u.EffectiveMessage.ID,
		})
	}
	if err != nil {
		utils.Logger.Sugar().Error(err)
		ctx.Reply(u, fmt.Sprintf("Error - %s", err.Error()), nil)
	}
	return dispatcher.EndGroups
}
