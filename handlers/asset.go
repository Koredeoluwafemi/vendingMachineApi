package handlers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"strconv"
)

func getUserID(c *fiber.Ctx) (uint, error) {
	if c.Locals("user") == nil {
		return 0, errors.New("unable to decode token")
	}
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	stringData := TransToString(claims["uid"])

	uid, err := strconv.Atoi(stringData)
	if err != nil {
		//log.Println(err)
		return 0, err
	}

	return uint(uid), nil
}

func getExpKey(c *fiber.Ctx) (int64, error) {
	if c.Locals("user") == nil {
		return 0, errors.New("unable to decode token")
	}
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	stringData := TransToString(claims["exp"])
	intData, err := strconv.Atoi(stringData)
	if err != nil {
		return 0, err
	}

	return int64(intData), nil
}

func getRole(c *fiber.Ctx) string {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	role := TransToString(claims["role"])

	return role
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

func check(c *fiber.Ctx, data interface{}, message string, status bool, code int) error {
	type ApiResponse struct {
		Status  bool        `json:"status"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	response := ApiResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	return c.Status(code).JSON(response)
}

func contain(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
