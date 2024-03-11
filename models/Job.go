package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	JobName      string         `json:"jobName"`
	SpecialistID uint           `json:"specialistID"`
	Frequencies  datatypes.JSON `json:"frequencies"`
}
