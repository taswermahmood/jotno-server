package models

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	SpecialistID uint   `json:"specialistID"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Stars        string `json:"stars"`
	Title        string `json:"title"`
	Body         string `json:"body"`
}
