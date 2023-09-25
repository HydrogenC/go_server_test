package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	gorm.Model
	Name     *string `json:"name"`
	Nickname *string `json:"nickname"`
	Age      *int    `json:"age"`
	// 0 for Male, 1 for Female
	Gender *bool `json:"gender"`
}

func getUsers(c *fiber.Ctx, db *gorm.DB) error {
	var users []User
	db.Find(&users)

	response, err := json.Marshal(users)
	if err != nil {
		return err
	}

	return c.SendString(string(response))
}

func querySingleUser(c *fiber.Ctx, db *gorm.DB) error {
	id, error := strconv.Atoi(c.Params("id", "-1"))
	if id < 0 || error != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Field 'id' isn't properly formed")
	}

	var user User
	res := db.First(&user, id)
	if res.Error != nil {
		return res.Error
	}

	response, err := json.Marshal(user)

	if err != nil {
		return err
	}
	return c.SendString(string(response))
}

func searchSingleUser(c *fiber.Ctx, db *gorm.DB) error {
	keyword := c.Query("search", "")
	if keyword == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Field 'search' isn't properly formed")
	}

	var users []User
	res := db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", keyword)).Find(&users)
	if res.Error != nil {
		return res.Error
	}

	response, err := json.Marshal(users)

	if err != nil {
		return err
	}
	return c.SendString(string(response))
}

func createUser(c *fiber.Ctx, db *gorm.DB) error {
	var user User
	json.Unmarshal(c.Body(), &user)

	if user.Name == nil || user.Nickname == nil || user.Age == nil || user.Gender == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Required field missing in request")
	}

	res := db.Create(&user)
	if res.Error!=nil{
		return res.Error
	}

	response, _ := json.Marshal(struct {
		Status string `json:"status"`
	}{"Success"})

	return c.SendString(string(response))
}

func removeUser(c *fiber.Ctx, db *gorm.DB) error{
	id, error := strconv.Atoi(c.Params("id", "-1"))
	if id < 0 || error != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Field 'id' isn't properly formed")
	}

	res:=db.Delete(&User{}, id)
	if res.Error!=nil{
		return res.Error
	}

	response, _ := json.Marshal(struct {
		Status string `json:"status"`
	}{"Success"})
	return c.SendString(string(response))
}

func main() {
	dsn := "host=localhost user=postgres password=admin dbname=gotest port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("Connection to database failed")
	}

	app := fiber.New()
	app.Get("/users", func(c *fiber.Ctx) error {
		return getUsers(c, db)
	})

	app.Post("/users", func(c *fiber.Ctx) error {
		return createUser(c, db)
	})

	app.Get("/users/:id<int>", func(c *fiber.Ctx) error {
		return querySingleUser(c, db)
	})

	app.Delete("/users/:id<int>", func(c *fiber.Ctx) error {
		return removeUser(c, db)
	})

	app.Get("/users/search", func(c *fiber.Ctx) error {
		return searchSingleUser(c, db)
	})

	log.Fatal(app.Listen(":8848"))
}
