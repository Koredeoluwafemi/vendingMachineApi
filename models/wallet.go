package models

import (
	"time"
)

type Wallet struct {
	ID        uint `gorm:"primary_key"`
	UserID    uint
	User      User
	Debit     int
	Credit    int
	Balance   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
