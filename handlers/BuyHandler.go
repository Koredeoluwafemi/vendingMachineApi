package handlers

import (
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"mvpmatch/database"
	"mvpmatch/models"
	"strings"
)

func getAllowedCoins() []int {
	return []int{5, 10, 20, 50, 100}
}

func getAllowedCoinsString() (string, error) {

	data := getAllowedCoins()
	s, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	str := strings.Trim(string(s), "[]")
	return str, nil
}

type depositInput struct {
	Coin int `json:"coin"`
}

func (s depositInput) Validate() error {
	valid := validation.ValidateStruct(&s,
		validation.Field(&s.Coin, validation.Required),
	)
	allowedCoins := getAllowedCoins()
	if !contain(allowedCoins, s.Coin) {
		allowedCoinsString, err := getAllowedCoinsString()
		if err != nil {
			return errors.New("invalid coin")
		}
		return errors.New("invalid coin, please supply one of these " + allowedCoinsString)
	}

	return valid
}

func Deposit(c *fiber.Ctx) error {

	var input depositInput
	db := database.DB

	userID, err := getUserID(c)
	if err != nil {
		return check(c, err, err.Error(), false, 401)
	}

	if err := c.BodyParser(&input); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	if err := input.Validate(); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	//set buyer deposit balance
	var buyer models.User
	db.First(&buyer, userID)

	balance := buyer.Deposit + input.Coin

	updateUser := db.Model(&models.User{})
	updateUser.Where(&models.User{ID: userID})
	rows := updateUser.Updates(models.User{Deposit: balance})

	if rows.RowsAffected == 0 {
		return check(c, "", "unable to save deposit", false, 401)
	}

	//log transaction in wallet
	wallet := models.Wallet{
		UserID:  userID,
		Credit:  input.Coin,
		Balance: balance,
	}

	//save coin
	var isCoinExist models.Coin
	rows = db.Where(&models.Coin{Denomination: input.Coin}).First(&isCoinExist)
	if rows.RowsAffected == 0 {
		//insert new coin
		coinModel := models.Coin{
			Denomination: input.Coin,
			Count:        1,
		}
		rows = db.Create(&coinModel)
	} else {
		//update coin count
		updateCoin := db.Model(&models.Coin{})
		updateCoin.Where("id = ?", isCoinExist.ID)
		rows = updateCoin.Updates(models.Coin{Count: isCoinExist.Count + 1})
	}

	db.Create(&wallet)

	return check(c, "", "deposit saved successfully!", true, 200)
}

type buyInput struct {
	ProductID uint `json:"product_id"`
	Amount    int  `json:"amount"`
}

func (s buyInput) Validate() error {
	valid := validation.ValidateStruct(&s,
		validation.Field(&s.ProductID, validation.Required),
	)
	db := database.DB
	var product models.Product
	rows := db.Where(&models.Product{ID: s.ProductID}).First(&product)
	if rows.RowsAffected == 0 {
		return errors.New("product_id is invalid")
	}

	if product.AmountAvailable < s.Amount {
		return errors.New("amount selected exceeds available amount")
	}

	return valid
}
func Buy(c *fiber.Ctx) error {

	var input buyInput
	db := database.DB

	userID, err := getUserID(c)
	if err != nil {
		return check(c, err, err.Error(), false, 401)
	}

	if err := c.BodyParser(&input); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	if err := input.Validate(); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	//get buyer's deposit balance
	var buyer models.User
	db.First(&buyer, userID)

	var product models.Product
	db.First(&product, input.ProductID)

	totalCost := product.Cost * input.Amount

	if totalCost > buyer.Deposit {
		return check(c, "", "insufficient deposit balance", false, 401)
	}

	change := buyer.Deposit - totalCost

	if change > 0 {
		if !isMultipleOf5(change) {
			return check(c, "", "Exact change or a multiple of 5 only", false, 400)
		}
	}

	//check if product has enough quantity to match request
	if product.AmountAvailable < input.Amount {
		return check(c, "", "Insufficient product quantity, please reduce the amount", false, 400)
	}

	//get available coins
	var availableCoins []models.Coin
	rows := db.Order("denomination desc").Find(&availableCoins)
	if rows.RowsAffected == 0 {
		return check(c, "", "no available coins", false, 400)
	}

	//log.Println(availableCoins)

	//log.Println(buyer.Deposit)
	//log.Println(totalCost)
	//log.Println(change)
	//log.Println(calculatedChange)

	var changeSlice []int

	if change > 0 {
		for _, item := range availableCoins {
			if change == item.Denomination {
				change = change - item.Denomination
				changeSlice = append(changeSlice, item.Denomination)
			}

			if change > item.Denomination {
				change = change - item.Denomination
				changeSlice = append(changeSlice, item.Denomination)
			}

			if change > item.Denomination {
				change = change - item.Denomination
				changeSlice = append(changeSlice, item.Denomination)
			}
		}

		if len(changeSlice) == 0 {
			return check(c, "","no coins available for change", false, 400)
		}

		calculatedChange := 0

		for _, eachChange := range changeSlice {
			calculatedChange = calculatedChange + eachChange
		}

		if calculatedChange < (buyer.Deposit - totalCost) {
			return check(c, "","unable to sell product, insufficient change", false, 400)
		}


		//process change for buyer
		for _, denomination := range changeSlice {
			//reduce denomination count by 1
			var getDenomination models.Coin
			newCount := getDenomination.Count - 1
			db.Where(&models.Coin{Denomination: denomination}).First(&getDenomination)
			updateCoin := db.Model(&models.Coin{})
			updateCoin.Where(&models.Coin{Denomination: denomination})
			updateCoin.Update("count",newCount)
		}

	} else {
		changeSlice = append(changeSlice, 0)
	}



	order := models.Order{
		ProductID: input.ProductID,
		UserID:    buyer.ID,
		Amount:    input.Amount,
	}

	rows = db.Create(&order)
	if rows.RowsAffected == 0 {
		return check(c, "", "unable to process order", false, 400)
	}

	//update product inventory
	productAmount := product.AmountAvailable - input.Amount
	updateProduct := db.Model(&models.Product{})
	updateProduct.Where(&models.Product{ID: input.ProductID})
	updateProduct.Update("amount_available", productAmount)


	//update user deposit
	upDeposit := db.Model(&models.User{})
	upDeposit.Where(&models.User{ID: buyer.ID})
	upDeposit.Update("deposit", change)

	amountSpent := product.Cost * input.Amount
	output := fiber.Map{
		"change":          changeSlice,
		"amount_spent":    amountSpent,
		"product":         product.ProductName,
		"number_of_units": input.Amount,
	}

	return check(c, output, "success", true, 200)
}

func isMultipleOf5(n int) bool {

	for n > 0 {
		n = n - 5
	}

	if n == 0 {
		return true
	}

	return false
}
