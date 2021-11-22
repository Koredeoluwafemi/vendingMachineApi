package routes

import (
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"mvpmatch/config"
	"mvpmatch/handlers"
	"mvpmatch/middleware"
)

func Routes(app *fiber.App) {

	jwtToken := jwtware.New(jwtware.Config{
		SigningKey: []byte(config.App.JWTKey),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == "Missing or malformed JWT" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Missing or malformed JWT", "status": false})
			} else {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid or expired JWT", "status": false})
			}
		},
	})

	route := app.Group("/v1")
	userRoutes(route, jwtToken)
}

func userRoutes(route fiber.Router, token fiber.Handler) {

	route.Post("user", handlers.AddUser)
	route.Get("user", handlers.GetUsers)
	route.Patch("user", token, handlers.EditUser)
	route.Delete("user", token, handlers.DeleteUser)

	route.Post("login", handlers.Login)
	route.Post("logout", token, handlers.Logout)
	route.Post("login/test", handlers.Logintest)

	route.Post("product", token, middleware.Seller, handlers.AddProduct)
	route.Get("product", handlers.GetProducts)
	route.Put("product", token, middleware.Seller, handlers.EditProduct)
	route.Delete("product", token, middleware.Seller, handlers.DeleteProduct)

	route.Post("deposit", token, middleware.Buyer, handlers.Deposit)
	route.Post("buy", token, middleware.Buyer, handlers.Buy)
	route.Patch("deposit/reset", token, handlers.ResetDeposit)

	route.Get("role", handlers.GetRole)
}
