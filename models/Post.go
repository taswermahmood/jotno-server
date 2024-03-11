package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	SpecialistID uint   `json:"specialistID"`
	Caption      string `json:"caption"`
	Media        string `json:"media"`
}
