package models

import (
	"time"
)

type Order struct {
	ID        uint `gorm:"primary_key"`
	ProductID uint
	Product   Product
	UserID    uint
	User      User
	Amount    int
	CreatedAt time.Time
	UpdatedAt time.Time
}
