package models

import (
	"gorm.io/gorm"
)

type JobPost struct {
	gorm.Model
	UserID        uint      `json:"userID"`
	JobType       string    `json:"jobType"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Wage          int       `json:"wage"`
	WageCurrency  string    `json:"wageCurrency"`
	WageFrequency string    `json:"wageFrequency"`
	DateTime      string    `json:"dateTime"`
	Comments      []Comment `json:"comments"`
}
