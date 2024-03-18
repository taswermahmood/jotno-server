package models

import "gorm.io/gorm"

type Booking struct {
	gorm.Model
	UserID       uint   `json:"userID"`
	SpecialistID uint   `json:"specialistID"`
	JobType      string `json:"jobType"`
	Active       bool   `json:"active"`
	Status       string `json:"status"`
	Frequency    string `json:"frequency"`
	Amount       int32  `json:"amount"`
	Currency     string `json:"currency"`
	Overdue      bool   `json:"overdue"`
	Bills        []Bill `json:"bills"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
}
