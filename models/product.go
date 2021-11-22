package models

import (
	"time"
)

type Product struct {
	ID              uint `gorm:"primary_key"`
	AmountAvailable int
	Cost            int
	ProductName     string
	SellerID        uint
	Seller          User
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
