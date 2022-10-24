package main

import (
	"net/http"
	"strconv"
	"time"

	"prom/app/mysql"

	"github.com/gofiber/fiber/v2"
)

// List Users
// @Summary List Users Service
// @Id list_users
// @version 1.0
// @produce application/json
// @Success 200 {object} []mysql.User
// @Router /v1/user [get]
// List Users Handler
func ListUsers(c *fiber.Ctx) error {
	time.Sleep(time.Second)
	userList := make([]*mysql.User, 0)
	res := userRepo.Db.Find(&userList)

	if res.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(res.Error)
	}

	return c.Status(http.StatusOK).JSON(userList)
}

// Get User
// @Summary Get User Service
// @Id get_user
// @version 1.0
// @produce application/json
// @Param id path int true "id"
// @Success 200 {object} mysql.User
// @Router /v1/user/{id} [get]
// Get User Handler
func GetUser(c *fiber.Ctx) error {

	uid := c.Params("id")
	user := &mysql.User{}
	res := userRepo.Db.Where("id = ?", uid).Find(user)

	if res.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(res.Error)
	}
	return c.Status(http.StatusOK).JSON(user)
}

// Create User
// @Summary Creates a User
// @Id create_user
// @version 1.0
// @produce application/json
// @Success 200 {object} mysql.User
// @Param name query string true "name"
// @Router /v1/user [put]
// Create User Handler
func CreateUser(c *fiber.Ctx) error {
	user := &mysql.User{
		Name: c.Query("name"),
	}

	res := userRepo.Db.Create(user)

	if res.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(res.Error)
	}
	return c.Status(http.StatusOK).JSON(user)
}

// Update User
// @Summary Update a User
// @Id update_user
// @version 1.0
// @produce application/json
// @Success 200 {object} mysql.User
// @Param id path string true "id"
// @Param name query string true "name"
// @Router /v1/user/{id} [post]
// Update User Handler
func UpdateUser(c *fiber.Ctx) error {
	uid := c.Params("id")
	user := &mysql.User{
		Name: c.Query("name"),
	}

	res := userRepo.Db.Where("id =string ?", uid).Updates(user)

	if res.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(res.Error)
	}

	// get user
	userRepo.Db.Where("id = ?", uid).Find(user)

	return c.Status(http.StatusOK).JSON(user)
}

// Delete User
// @Summary Delete a User
// @Id delete_user
// @version 1.0
// @produce application/json
// @Success 200 {string} string "success"
// @Param id path string true "id"
// @Router /v1/user/{id} [delete]
// Delete User Handler
func DeleteUser(c *fiber.Ctx) error {
	uid, _ := strconv.Atoi(c.Params("id"))
	res := userRepo.Db.Delete(&mysql.User{
		Id: uid,
	})

	if res.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(res.Error)
	}
	return c.Status(http.StatusOK).SendString("success")
}
