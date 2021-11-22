package database

import (
	"gorm.io/gorm"
	"mvpmatch/config"
	"mvpmatch/models"
)

func seed(db *gorm.DB) {
	roleSeeder(db)
}

func roleSeeder(db *gorm.DB) {
	name := config.Role.Buyer
	status := models.Role{Name: name}
	db.Where(status).FirstOrCreate(&status)

	name = config.Role.Seller
	status = models.Role{Name: name}
	db.Where(status).FirstOrCreate(&status)
}
