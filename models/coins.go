package models

import (
	"time"
)

type Coin struct {
	ID           uint `gorm:"primary_key"`
	Denomination int
	Count        int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
