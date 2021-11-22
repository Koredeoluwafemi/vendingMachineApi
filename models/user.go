package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint `gorm:"primary_key"`
	Username  string
	Password  string
	Deposit   int
	RoleID    uint
	Role      Role
	Deleted gorm.DeletedAt
	CreatedAt time.Time
	UpdatedAt time.Time
}
