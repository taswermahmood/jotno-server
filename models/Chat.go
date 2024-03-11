package models

import "gorm.io/gorm"

type Chat struct {
	gorm.Model
	UserID       uint      `json:"userID"`
	SpecialistID uint      `json:"specialistID"`
	Messages     []Message `json:"messages"`
}
