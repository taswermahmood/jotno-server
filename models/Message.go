package models

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	ChatID     uint   `json:"chatID"`
	SenderID   uint   `json:"senderID"`
	ReceiverID uint   `json:"receiverID"`
	Text       string `json:"text"`
}
