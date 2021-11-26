package tests

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/stretchr/testify/assert"
	"log"
	"mvpmatch/config"
	"mvpmatch/database"
	"mvpmatch/handlers"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuyRoute(t *testing.T) {

	longLivedBuyertoken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mzc4OTc3OTgsIm5hbWUiOiJidXllciIsInJpZCI6MSwicm9sZSI6ImJ1eWVyIiwidWlkIjo2fQ.WqbpORAgMkWDT2qkn2KMZwDxkYcmG6Ef8ibiqbtkByY"
	type payloadStruct struct {
		ProductID uint `json:"product_id"`
		Amount    int  `json:"amount"`
	}
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		expectedCode int    // expected HTTP status code
		payload      payloadStruct
		token        string
	}{
		{
			description:  "Test: is endpoint secured, get http status 400 ",
			route:        "/buy",
			expectedCode: 400,
			payload: payloadStruct{
				ProductID: 34,
				Amount:    0,
			},
			token: "",
		},
		{
			description:  "Test: buy invalid product, get HTTP status 400",
			route:        "/buy",
			expectedCode: 400,
			payload: payloadStruct{
				ProductID: 399874,
				Amount:    0,
			},
			token: longLivedBuyertoken,
		},
		{
			description:  "Test: test for insufficient balance, get HTTP status 400",
			route:        "/buy",
			expectedCode: 400,
			payload: payloadStruct{
				ProductID: 34,
				Amount:    2000000000000,
			},
			token: longLivedBuyertoken,
		},
		{
			description:  "Test: buy valid product but using zero amount, get HTTP status 400",
			route:        "/buy",
			expectedCode: 400,
			payload: payloadStruct{
				ProductID: 4,
				Amount:    0,
			},
			token: longLivedBuyertoken,
		},
	}

	// Define Fiber app.
	app := fiber.New()
	database.Start()
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
	app.Post("buy", jwtToken, handlers.Buy)

	// Iterate through test single test cases
	for _, test := range tests {
		// Create a new http request with the route from the test case
		payload, err := json.Marshal(test.payload)
		if err != nil {
			panic(err)
		}

		req := httptest.NewRequest(http.MethodPost, test.route, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+test.token)

		// Perform the request plain with the app,
		resp, err := app.Test(req, -1)
		if err != nil {
			log.Println(err)
		}

		// Verify, if the status code is as expected
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.description)
	}
}
