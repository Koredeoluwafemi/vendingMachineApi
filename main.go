package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"mvpmatch/database"
	"mvpmatch/routes"
	"os"
)

func main() {

	//get root directory
	path, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	resourcesPath := path + "/" + "resources"

	app := fiber.New(fiber.Config{})

	//start database
	database.Start()
	database.Migrate()

	routes.Routes(app)

	app.Static("/", resourcesPath)

	port := "3000"

	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}
