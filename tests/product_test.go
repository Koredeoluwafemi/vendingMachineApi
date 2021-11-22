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
	"mvpmatch/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductRoute(t *testing.T) {

	longLivedSellerToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mzc5MzkxNTcsIm5hbWUiOiJzZWxsZXIxIiwicmlkIjoyLCJyb2xlIjoic2VsbGVyIiwidWlkIjo1fQ.LbNm77Bnf4mSfqpRuvVGeOb-nbFV1lh927kopU70gL0"
	// Define a structure for specifying input and output data
	// of a single test case
	type payloadStruct struct {
		AmountAvailable int    `json:"amount_available"`
		Cost            int    `json:"cost"`
		ProductName     string `json:"product_name"`
	}
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		expectedCode int    // expected HTTP status code
		payload      payloadStruct
		token string
	}{
		// First test case
		{
			description:  "Test: is endpoint secured, get http status 400 ",
			route:        "/product",
			expectedCode: 400,
			payload: payloadStruct{
				AmountAvailable: 34,
				Cost: 0,
				ProductName: "",
			},
			token: "",
		},
		{
			description:  "Test: add product, get HTTP status 201",
			route:        "/product",
			expectedCode: 201,
			payload: payloadStruct{
				AmountAvailable: 34,
				Cost: 20,
				ProductName: "product_name",
			},
			token: longLivedSellerToken,
		},
		{
			description:  "Test: test for duplicate product name, get HTTP status 400",
			route:        "/product",
			expectedCode: 400,
			payload: payloadStruct{
				AmountAvailable: 34,
				Cost: 20,
				ProductName: "product_name",
			},
			token: longLivedSellerToken,
		},
		{
			description:  "Test: add product with an invalid cost, get HTTP status 400",
			route:        "/product",
			expectedCode: 400,
			payload: payloadStruct{
				AmountAvailable: 34,
				Cost: 13,
				ProductName: "prodd",
			},
			token: longLivedSellerToken,
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
	app.Post("product", jwtToken, middleware.Seller, handlers.AddProduct)


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
		// the second argument is a request latency
		// (set to -1 for no latency)
		resp, err := app.Test(req, -1)
		if err != nil {
			log.Println(err)
		}

		// Verify, if the status code is as expected
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.description)
	}
}
