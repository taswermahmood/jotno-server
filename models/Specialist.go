package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Specialist struct {
	gorm.Model
	FirstName           string         `json:"firstName"`
	LastName            string         `json:"lastName"`
	Email               string         `json:"email"`
	Password            string         `json:"password"`
	CountryCode         string         `json:"countryCode"`
	CallingCode         string         `json:"callingCode"`
	PhoneNumber         string         `json:"phoneNumber"`
	Avatar              string         `json:"avatar"`
	Images              datatypes.JSON `json:"images"`
	IdCard              string         `json:"idCard"`
	Address             string         `json:"address"`
	City                string         `json:"city"`
	Lat                 float32        `json:"lat"`
	Lon                 float32        `json:"lon"`
	Experience          int            `json:"experience"`
	Stars               int            `json:"stars"`
	About               string         `json:"about"`
	Verified            bool           `json:"verified"`
	Jobs                []Job          `json:"jobs"`
	Reviews             []Review       `json:"reviews"`
	Posts               []Post         `json:"posts"`
	PushTokens          datatypes.JSON `json:"pushTokens"`
	AllowsNotifications *bool          `json:"allowsNotifications"`
}
