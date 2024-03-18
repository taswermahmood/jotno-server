package models

import "gorm.io/gorm"

type Bill struct {
	gorm.Model
	BookingID uint   `json:"bookingID"`
	Paid      bool   `json:"paid"`
	Received  bool   `json:"received"`
	Complete  bool   `json:"complete"`
	Amount    int32  `json:"amount"`
	Currency  string `json:"currency"`
}
