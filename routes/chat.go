package routes

import (
	"errors"
	"jotno-server/models"
	"jotno-server/storage"
	"jotno-server/utils"
	"sort"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func CreateChat(ctx iris.Context) {
	var req CreateChatInput

	err := ctx.ReadJSON(&req)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}

	var prevChat models.Chat
	chatExists := storage.DB.
		Where("user_id = ? AND specialist_id = ? ", req.UserID, req.SpecialistID).
		Find(&prevChat)

	if chatExists.Error != nil {
		utils.InternalServerError(ctx)
		return
	}

	if chatExists.RowsAffected > 0 {
		ctx.StatusCode(iris.StatusConflict)
		ctx.Text("Chat already exists")
		return
	}

	var messages []models.Message
	messages = append(messages, models.Message{
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		Text:       req.Text,
	})

	chat := models.Chat{
		UserID:       req.UserID,
		SpecialistID: req.SpecialistID,
		Messages:     messages,
	}

	storage.DB.Create(&chat)

	ctx.JSON(chat)
}

func GetChatByUserAndSpecialistID(ctx iris.Context) {
	var req GetChatInput
	err := ctx.ReadJSON(&req)
	
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}
	
	result, err := getChatResultsByUserIDAndSpecialistID(req.UserID, req.SpecialistID, ctx)
	if err != nil {
		return
	}
	var messages []models.Message
	messagesQuery := storage.DB.Where("chat_id = ?", result.ID).Order("created_at DESC").Find(&messages)

	if messagesQuery.Error != nil {
		utils.InternalServerError(ctx)
		return
	}

	result.Messages = messages

	ctx.JSON(result)
}

func GetChatByID(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("chatId")

	result, err := getChatResult(id, ctx)

	if err != nil {
		return
	}

	var messages []models.Message
	messagesQuery := storage.DB.Where("chat_id = ?", id).Order("created_at DESC").Find(&messages)

	if messagesQuery.Error != nil {
		utils.InternalServerError(ctx)
		return
	}

	result.Messages = messages

	ctx.JSON(result)
}

func GetChatsByUserID(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("id")

	results, err := getChatResultsByUserID(id, ctx)

	if err != nil {
		return
	}

	var chatIDs []uint
	for _, chat := range results {
		chatIDs = append(chatIDs, chat.ID)
	}

	var messages []models.Message

	messagesQuery := storage.DB.Raw(`
		SELECT messages.* 
		FROM messages
		INNER JOIN (
			SELECT chat_id, MAX(created_at) AS created_at
			FROM messages
			WHERE chat_id IN ? 
			GROUP BY chat_id
		) AS recentMessages
		ON messages.chat_id = recentMessages.chat_id 
		AND messages.created_at = recentMessages.created_at`, chatIDs).Scan(&messages)

	messageMap := make(map[uint][]models.Message)
	for _, message := range messages {
		messageSlice := []models.Message{message}
		messageMap[message.ChatID] = messageSlice
	}

	for index, chat := range results {
		results[index].Messages = messageMap[chat.ID]
	}

	if messagesQuery.Error != nil {
		utils.InternalServerError(ctx)
		return
	}

	sort.Slice(results, func(i int, j int) bool {
		return results[i].Messages[0].CreatedAt.After(results[j].Messages[0].CreatedAt)
	})

	ctx.JSON(results)
}

func getChatResult(id string, ctx iris.Context) (ChatResult, error) {
	var result ChatResult
	resultQuery := storage.DB.Table("chats").
		Select(`chats.*,
		 specialists.first_name as specialist_first_name, specialists.last_name as specialist_last_name, specialists.avatar as specialist_avatar,
		 users.first_name as user_first_name, users.last_name as user_last_name, users.avatar as user_avatar`).
		Joins("INNER JOIN specialists on chats.specialist_id = specialists.id").
		Joins("INNER JOIN users on chats.user_id = users.id").
		Where("chats.id = ?", id).
		Scan(&result)

	if resultQuery.Error != nil {
		utils.InternalServerError(ctx)
		return result, resultQuery.Error
	}

	if resultQuery.RowsAffected == 0 {
		utils.CreateNotFound(ctx)
		return result, errors.New("result not found")
	}

	return result, nil
}

func getChatResultsByUserIDAndSpecialistID(userId uint, specialistId uint, ctx iris.Context) (ChatResult, error) {
	var result ChatResult
	resultQuery := storage.DB.Table("chats").
		Select(`chats.*,
		 specialists.first_name as specialist_first_name, specialists.last_name as specialist_last_name, specialists.avatar as specialist_avatar,
		 users.first_name as user_first_name, users.last_name as user_last_name, users.avatar as user_avatar`).
		Joins("INNER JOIN specialists on chats.specialist_id = specialists.id").
		Joins("INNER JOIN users on chats.user_id = users.id").
		Where("chats.user_id = ? AND chats.specialist_id = ? ", userId, specialistId).
		Scan(&result)

	if resultQuery.Error != nil {
		utils.InternalServerError(ctx)
		return result, resultQuery.Error
	}

	if resultQuery.RowsAffected == 0 {
		utils.CreateNotFound(ctx)
		return result, errors.New("result not found")
	}

	return result, nil
}

func getChatResultsByUserID(id string, ctx iris.Context) ([]ChatResult, error) {
	var result []ChatResult
	resultQuery := storage.DB.Table("chats").
		Select(`chats.*,
		 specialists.first_name as specialist_first_name, specialists.last_name as specialist_last_name, specialists.avatar as specialist_avatar,
		 users.first_name as user_first_name, users.last_name as user_last_name, users.avatar as user_avatar`).
		Joins("INNER JOIN specialists on chats.specialist_id = specialists.id").
		Joins("INNER JOIN users on chats.user_id = users.id").
		Where("chats.user_id = ?", id).Or("chats.specialist_id = ?", id).
		Scan(&result)

	if resultQuery.Error != nil {
		utils.InternalServerError(ctx)
		return result, resultQuery.Error
	}

	if resultQuery.RowsAffected == 0 {
		utils.CreateNotFound(ctx)
		return result, errors.New("result not found")
	}

	return result, nil
}

type ChatResult struct {
	// Chat
	gorm.Model
	UserID       uint `json:"userID"`
	SpecialistID uint `json:"specialistID"`
	// Specialist
	SpecialistFirstName string `json:"specialistFirstName"`
	SpecialistLastName  string `json:"specialistLastName"`
	SpecialistAvatar    string `json:"specialistAvatar"`
	// User
	UserFirstName string `json:"userFirstName"`
	UserLastName  string `json:"userLastName"`
	UserAvatar    string `json:"userAvatar"`
	// Message
	Messages []models.Message `gorm:"foreignKey:ID" json:"messages"`
}

type CreateChatInput struct {
	UserID       uint   `json:"userID" validate:"required"`
	SpecialistID uint   `json:"specialistID" validate:"required"`
	SenderID     uint   `json:"senderID" validate:"required"`
	ReceiverID   uint   `json:"receiverID" validate:"required"`
	Text         string `json:"text" validate:"required,lt=5000"`
}

type GetChatInput struct {
	UserID       uint `json:"userID" validate:"required"`
	SpecialistID uint `json:"specialistID" validate:"required"`
}
