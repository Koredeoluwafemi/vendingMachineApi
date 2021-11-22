package database

import (
	"log"
	"mvpmatch/models"
)

func Migrate() {
	// Migrate the schema
	db := DB
	err := db.AutoMigrate(
		&models.Order{},
		&models.Product{},
		&models.User{},
		&models.Role{},
		&models.Wallet{},
		&models.Coin{},
	)

	if err != nil {
		log.Println(err)
	}

	seed(db)
}
