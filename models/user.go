package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"-" gorm:"not null"`
	Balance  int    `json:"balance" gorm:"default:0"`
	Referrer *int   `json:"referrer,omitempty" gorm:"default:NULL"`
}

type ReferralRequest struct {
	ReferrerID int `json:"referrer_id"`
}
