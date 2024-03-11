package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName           string         `json:"firstName"`
	LastName            string         `json:"lastName"`
	Email               string         `json:"email"`
	Password            string         `json:"password"`
	CountryCode         string         `json:"countryCode"`
	CallingCode         string         `json:"callingCode"`
	PhoneNumber         string         `json:"phoneNumber"`
	Address             string         `json:"address"`
	City                string         `json:"city"`
	Lat                 float32        `json:"lat"`
	Lon                 float32        `json:"lon"`
	Avatar              string         `json:"avatar"`
	SocialLogin         bool           `json:"socialLogin"`
	SocialProvider      string         `json:"socialProvider"`
	JobPosts            []JobPost      `json:"jobPosts"`
	Favorited           datatypes.JSON `json:"favorited"`
	PushTokens          datatypes.JSON `json:"pushTokens"`
	AllowsNotifications *bool          `json:"allowsNotifications"`
}
