package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"mvpmatch/config"
	"mvpmatch/database"
	"strconv"
	"strings"
)

func Seller(c *fiber.Ctx) error {

	if c.Locals("user") == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token access required", "status": false})
	}
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	blacklist, err := checkBlacklist(c, claims)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token logged out, Please login afresh", "status": false})
	}

	if blacklist {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token logged out, Please login afresh", "status": false})
	}

	slug := fmt.Sprintf("%v", claims["role"])
	if slug == config.Role.Seller {
		return c.Next()
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": config.Role.Seller + " access required", "status": false})
	}
}
func Buyer(c *fiber.Ctx) error {

	if c.Locals("user") == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token access required", "status": false})
	}
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	blacklist, err := checkBlacklist(c, claims)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token logged out, Please login afresh", "status": false})
	}

	if blacklist {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Token logged out, Please login afresh", "status": false})
	}

	slug := fmt.Sprintf("%v", claims["role"])
	if slug == config.Role.Buyer {
		return c.Next()
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": config.Role.Buyer + " access required", "status": false})
	}
}

func checkBlacklist(c *fiber.Ctx, claims jwt.MapClaims) (bool, error) {

	client, err :=  database.ConnectRedis()
	if err != nil {
		return false, errors.New("unable to verify token")
	}
	defer client.Close()

	stringData := TransToString(claims["uid"])
	token := client.Get(context.TODO(), stringData)

	if token.Val() == "" {
		return false, nil
	}

	tokenHeader := strings.Split(c.Get("Authorization"), " ")
	if len(tokenHeader) != 2 {
		return true, errors.New("unable to get token")
	}
	jwtToken := tokenHeader[1]

	if jwtToken == token.Val() {
		return true, nil
	}

	return false, nil
}

func TransToString(data interface{}) (res string) {
	switch v := data.(type) {
	case float64:
		res = strconv.FormatFloat(data.(float64), 'f', 0, 64)
	case float32:
		res = strconv.FormatFloat(float64(data.(float32)), 'f', 6, 32)
	case int:
		res = strconv.FormatInt(int64(data.(int)), 10)
	case int64:
		res = strconv.FormatInt(data.(int64), 10)
	case uint:
		res = strconv.FormatUint(uint64(data.(uint)), 10)
	case uint64:
		res = strconv.FormatUint(data.(uint64), 10)
	case uint32:
		res = strconv.FormatUint(uint64(data.(uint32)), 10)
	case json.Number:
		res = data.(json.Number).String()
	case string:
		res = data.(string)
	case []byte:
		res = string(v)
	default:
		res = ""
	}
	return
}

