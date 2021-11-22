package handlers

import (
	"context"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
	"mvpmatch/config"
	"mvpmatch/database"
	"mvpmatch/models"
	"strconv"
	"strings"
	"time"
)

type addUserInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RoleID   uint   `json:"role_id"`
}
func (s addUserInput) Validate() error {
	valid := validation.ValidateStruct(&s,
		validation.Field(&s.Username, validation.Required),
		validation.Field(&s.Password, validation.Required),
		validation.Field(&s.RoleID, validation.Required),
	)

	db := database.DB
	rows := db.Where(&models.Role{ID: s.RoleID}).First(&models.Role{})
	if rows.RowsAffected == 0 {
		return errors.New("role_id is invalid")
	}

	rows = db.Where(&models.User{Username: s.Username}).First(&models.User{})
	if rows.RowsAffected == 1 {
		return errors.New("username is not available, use another!")
	}

	return valid
}
func AddUser(c *fiber.Ctx) error {

	var input addUserInput
	db := database.DB

	if err := c.BodyParser(&input); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	if err := input.Validate(); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	//create password hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return check(c, "", "Unable to encrypt password", false, 500)
	}

	user := models.User{
		Username: input.Username,
		Password: string(passwordHash),
		//Deposit:   "",
		RoleID: input.RoleID,
	}

	rows := db.Create(&user)
	if rows.RowsAffected == 0 {
		return check(c, "", "Unable to create user", false, 400)
	}

	db.Where("id = ?", user.ID).Preload(clause.Associations).First(&user)

	output := fiber.Map{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role.Name,
	}
	return check(c, output, "user created successfully", true, 201)
}
func GetUsers(c *fiber.Ctx) error {

	db := database.DB

	var users []models.User
	rows := db.Preload(clause.Associations).Find(&users)
	if rows.RowsAffected == 0 {
		empty := make([]string, 0)
		return check(c, empty, "no records found", true, 200)
	}

	type list struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	var allResult []list
	for _, item := range users {
		result := list{
			ID:       item.ID,
			Username: item.Username,
			Role:     item.Role.Name,
		}

		allResult = append(allResult, result)
	}

	return check(c, allResult, "users", true, 200)
}

type editUserInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
func (s editUserInput) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Username, validation.Required),
	)
}
func EditUser(c *fiber.Ctx) error {

	var input editUserInput
	db := database.DB

	userID, err := getUserID(c)
	if err != nil {
		return check(c, err, err.Error(), false, 401)
	}

	if err = c.BodyParser(&input); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	if err = input.Validate(); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	var user models.User
	rows := db.Where(&models.User{Username: input.Username}).First(&user)
	if rows.RowsAffected == 1 {
		if user.ID != userID {
			return check(c, "", "username is not available, use another!", false, 400)
		}
	}

	updateUser := db.Model(&models.User{})
	updateUser.Where(&models.User{ID: userID})


	if input.Password != "" {
		//create password hash
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return check(c, "", "Unable to encrypt password", false, 500)
		}

		rows = updateUser.Updates(models.User{Password: string(passwordHash), Username: input.Username})
	} else {
		rows = updateUser.Update("username", input.Username)
	}

	if rows.RowsAffected == 0 {
		return check(c, "", "unable to update user", false, 400)
	}

	//var user models.User
	db.Where("id = ?", userID).Preload(clause.Associations).First(&user)

	output := fiber.Map{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role.Name,
	}
	return check(c, output, "user edited successfully", true, 200)
}

func ResetDeposit(c *fiber.Ctx) error {

	//var input editUserInput
	db := database.DB

	userID, err := getUserID(c)
	if err != nil {
		return check(c, err, err.Error(), false, 401)
	}

	//update user deposit
	updateUser := db.Model(&models.User{})
	updateUser.Where(&models.User{ID: userID})
	rows := updateUser.Update("deposit", 0)
	if rows.RowsAffected == 0 {
		return check(c, "", "unable to reset deposit", false, 400)
	}

	var user models.User
	db.Where("id = ?", userID).Preload(clause.Associations).First(&user)

	output := fiber.Map{
		"username": user.Username,
		"role":     user.Role.Name,
		"deposit":  user.Deposit,
	}
	return check(c, output, "user deposit reset successful", true, 200)
}
func DeleteUser(c *fiber.Ctx) error {

	db := database.DB

	userID, err := getUserID(c)
	if err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	row := db.Delete(&models.User{ID: userID})
	if row.RowsAffected == 0 {
		return check(c, "", "unable to delete user", false, 400)
	}

	return check(c, "", "user deleted successfully!", true, 200)
}

func GetRole(c *fiber.Ctx) error {

	db := database.DB

	var roles []models.Role

	rows := db.Find(&roles)
	if rows.RowsAffected == 0 {
		empty := make([]string, 0)
		return check(c, empty, "no record found", false, 400)
	}

	type list struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	var allResult []list
	for _, item := range roles {
		result := list{
			ID:   item.ID,
			Name: item.Name,
		}

		allResult = append(allResult, result)
	}

	return check(c, allResult, "roles", true, 200)
}
type loginBody struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
func (s loginBody) Validate() error {
	var data error
	data = validation.ValidateStruct(&s,
		validation.Field(&s.Username, validation.Required),
		validation.Field(&s.Password, validation.Required),
	)
	db := database.DB
	rows := db.Where("username = ?", s.Username).First(&models.User{})
	if rows.RowsAffected == 0 {
		return errors.New("Username is invalid")
	}
	return data
}

func Login(c *fiber.Ctx) error {

	var input loginBody
	if err := c.BodyParser(&input); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	if err := input.Validate(); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	db := database.DB
	user := models.User{}

	db.Where("username = ?", input.Username).Preload(clause.Associations).First(&user)

	//user has been verified
	hash := []byte(user.Password)
	if err := bcrypt.CompareHashAndPassword(hash, []byte(input.Password)); err != nil {
		// TODO: Properly handle error
		return check(c, "", "Unable to login, credentials wrong", false, 401)
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["rid"] = user.RoleID
	claims["name"] = user.Username
	claims["role"] = user.Role.Name
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	// Generate encoded token and send it as response.
	tokenString, err := token.SignedString([]byte(config.App.JWTKey))
	if err != nil {
		return check(c, "", "Unable to generate token", false, 500)
	}

	//compose return message struct
	result := fiber.Map{
		"username": user.Username,
		"token":    tokenString,
	}
	return check(c, result, "success", true, 200)
}
func Logintest(c *fiber.Ctx) error {

	var input loginBody
	if err := c.BodyParser(&input); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	if err := input.Validate(); err != nil {
		return check(c, err, err.Error(), false, 400)
	}

	db := database.DB
	user := models.User{}

	db.Where("username = ?", input.Username).Preload(clause.Associations).First(&user)

	//user has been verified
	hash := []byte(user.Password)
	if err := bcrypt.CompareHashAndPassword(hash, []byte(input.Password)); err != nil {
		// TODO: Properly handle error
		return check(c, "", "Unable to login, credentials wrong", false, 401)
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["rid"] = user.RoleID
	claims["name"] = user.Username
	claims["role"] = user.Role.Name
	claims["exp"] = time.Now().Add(time.Hour * 100).Unix()

	// Generate encoded token and send it as response.
	tokenString, err := token.SignedString([]byte(config.App.JWTKey))
	if err != nil {
		return check(c, "", "Unable to generate token", false, 500)
	}

	//compose return message struct
	result := fiber.Map{
		"username": user.Username,
		"token":    tokenString,
	}
	return check(c, result, "success", true, 200)
}

func Logout(c *fiber.Ctx) error {

	tokenHeader := strings.Split(c.Get("Authorization"), " ")
	if len(tokenHeader) != 2 {
		return check(c,"","action invalid", false, 400)
	}
	token := tokenHeader[1]

	userID, err := getUserID(c)
	if err != nil {
		return check(c, err, err.Error(), false, 401)
	}



	//client := redis.NewClient(&redis.Options{
	//	Addr:     "localhost:6379", // host:port of the redis server
	//	Password: "", // no password set
	//	DB:       0,  // use default DB
	//})

	client, err := database.ConnectRedis()
	if err != nil {
		return check(c, err, err.Error(), false, 400)
	}
	defer client.Close()

	ctx := context.TODO()
	userIDKey := strconv.Itoa(int(userID))

	//expKey, err := getExpKey(c)
	//if err != nil {
	//	return check(c, err, err.Error(), false, 401)
	//}
	//t := time.Unix(expKey, 0)

	client.Set(ctx, userIDKey, token, 0)


	return check(c, "", "success", true, 200)
}
