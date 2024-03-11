package models

import (
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	JobPostID    uint   `json:"jobPostID"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Image        string `json:"image"`
	SpecialistID uint   `json:"specialistID"`
	Comment      string `json:"comment"`
}
