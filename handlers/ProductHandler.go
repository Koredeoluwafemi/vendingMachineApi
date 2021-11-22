package handlers

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"gorm.io/gorm/clause"
	"mvpmatch/database"
	"mvpmatch/models"
)

type addProductInput struct {
	AmountAvailable int    `json:"amount_available"`
	Cost            int    `json:"cost"`
	ProductName     string `json:"product_name"`
}

func (s addProductInput) Validate() error {
	valid := validation.ValidateStruct(&s,
		validation.Field(&s.AmountAvailable, validation.Required),
		validation.Field(&s.Cost, validation.Required),
		validation.Field(&s.ProductName, validation.Required),
	)

	db := database.DB
	rows := db.Where(&models.Product{ProductName: s.ProductName}).First(&models.Product{})
	if rows.RowsAffected == 1 {
		return errors.New("product_name already exists!")
	}

	//cost must be a multiple of 5
	if !isMultipleOf5(s.Cost) {
		return errors.New("cost must be a multiple of 5")
	}

	if s.AmountAvailable == 0 {
		return errors.New("amount_available cannot be lesser than 1")
	}

	return valid
}
func AddProduct(c *fiber.Ctx) error {

	var input addProductInput
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

	var user models.User
	row := db.First(&user, userID)
	if row.RowsAffected == 0 {
		return check(c,"","account no longer valid", false, 400)
	}

	product := models.Product{
		AmountAvailable: input.AmountAvailable,
		Cost:            input.Cost,
		ProductName:     input.ProductName,
		SellerID:        userID,
	}

	row = db.Create(&product)
	if row.RowsAffected == 0 {
		return check(c, "", "unable to add product", false, 400)
	}

	output := fiber.Map{
		"product_id": product.ID,
		"name":             input.ProductName,
		"amount_available": input.AmountAvailable,
		"cost":             input.Cost,
	}

	return check(c, output, "product created successfully", true, 201)
}

func GetProducts(c *fiber.Ctx) error {

	db := database.DB

	var products []models.Product
	rows := db.Preload(clause.Associations).Find(&products)
	if rows.RowsAffected == 0 {
		empty := make([]string, 0)
		return check(c, empty, "no records found", true, 200)
	}

	type list struct {
		ID              uint    `json:"id"`
		ProductName     string  `json:"product_name"`
		AmountAvailable int     `json:"amount_available"`
		Seller          string  `json:"seller"`
		Cost            int `json:"cost"`
	}

	var allResult []list
	for _, item := range products {
		result := list{
			ID:              item.ID,
			ProductName:     item.ProductName,
			AmountAvailable: item.AmountAvailable,
			Seller:          item.Seller.Username,
			Cost:            item.Cost,
		}

		allResult = append(allResult, result)
	}

	return check(c, allResult, "products", true, 200)
}

type editProductInput struct {
	ProductID       uint   `json:"product_id"`
	AmountAvailable int    `json:"amount_available"`
	Cost            int    `json:"cost"`
	ProductName     string `json:"product_name"`
}

func (s editProductInput) Validate() error {
	valid := validation.ValidateStruct(&s,
		validation.Field(&s.ProductID, validation.Required),
		validation.Field(&s.AmountAvailable, validation.Required),
		validation.Field(&s.Cost, validation.Required),
		validation.Field(&s.ProductName, validation.Required),
	)

	db := database.DB
	rows := db.Where(&models.Product{ID: s.ProductID}).First(&models.Product{})
	if rows.RowsAffected == 0 {
		return errors.New("product_id is invalid")
	}

	//cost must be a multiple of 5
	if !isMultipleOf5(s.Cost) {
		return errors.New("cost must be a multiple of 5")
	}

	return valid
}
func EditProduct(c *fiber.Ctx) error {

	var input editProductInput
	db := database.DB

	if err := c.BodyParser(&input); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	if err := input.Validate(); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	sellerID, err := getUserID(c)
	if err != nil {
		return check(c, err, err.Error(), false, 401)
	}

	//check if user owns product
	var isSellerProduct models.Product
	rows := db.Where(&models.Product{ID: input.ProductID, SellerID: sellerID}).First(&isSellerProduct)
	if rows.RowsAffected == 0 {
		return check(c, "", "permission denied!", false, 401)
	}

	//check name validity
	var nameExist models.Product
	rows = db.Where("product_name = ?", input.ProductName).First(&nameExist)
	if rows.RowsAffected == 1 {
		if nameExist.ID != input.ProductID {
			return check(c, "", "product_name exists!", false, 400)
		}
	}

	updateProduct := db.Model(&models.Product{})
	updateProduct.Where(&models.Product{ID: input.ProductID})
	rows = updateProduct.Updates(models.Product{
		AmountAvailable: input.AmountAvailable,
		Cost:            input.Cost,
		ProductName:     input.ProductName,
	})
	if rows.RowsAffected == 0 {
		return check(c, "", "unable to update product", false, 400)
	}

	var product models.Product
	db.Where("id = ?", input.ProductID).Preload(clause.Associations).First(&product)

	output := fiber.Map{
		"amount_available": product.AmountAvailable,
		"name":             product.ProductName,
		"cost":             product.Cost,
	}
	return check(c, output, "product edited successfully", true, 200)
}

type delProductInput struct {
	ProductID uint `json:"product_id"`
}

func (s delProductInput) Validate() error {
	valid := validation.ValidateStruct(&s,
		validation.Field(&s.ProductID, validation.Required),
	)

	db := database.DB
	rows := db.Where(&models.Product{ID: s.ProductID}).First(&models.Product{})
	if rows.RowsAffected == 0 {
		return errors.New("product_id is invalid")
	}

	return valid
}
func DeleteProduct(c *fiber.Ctx) error {

	var input delProductInput
	db := database.DB

	if err := c.BodyParser(&input); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	if err := input.Validate(); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	sellerID, err := getUserID(c)
	if err != nil {
		return check(c, err, err.Error(), false, 401)
	}

	//check if user owns product
	var isSellerProduct models.Product
	rows := db.Where(&models.Product{ID: input.ProductID, SellerID: sellerID}).First(&isSellerProduct)
	if rows.RowsAffected == 0 {
		return check(c, "", "permission denied!", false, 401)
	}

	row := db.Delete(&models.Product{ID: input.ProductID})
	if row.RowsAffected == 0 {
		return check(c, "", "unable to delete product", false, 400)
	}

	return check(c, "", "product deleted successfully!", true, 200)
}
