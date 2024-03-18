package models

import "gorm.io/gorm"

type Chat struct {
	gorm.Model
	UserID       uint      `json:"userID"`
	JobId        uint      `json:"jobId"`
	SpecialistID uint      `json:"specialistID"`
	Messages     []Message `json:"messages"`
}
