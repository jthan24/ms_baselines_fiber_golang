package main

import (
	"context"
	"errors"
	"net/http"
	"prom/app/config"
	"prom/app/db"
	"prom/core/domain/repository"
	"prom/core/usecases"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer(config.GetConfig().ServiceName)
var once sync.Once
var logger *otelzap.Logger

func Logger(ctx context.Context) otelzap.LoggerWithCtx {
	once.Do(func() {
		l, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		logger = otelzap.New(l)
	})
	return logger.Ctx(ctx)
}

// List Users
// @Summary List Users Service
// @Id list_users
// @version 1.0
// @produce application/json
// @Success 200 {object} []db.User
// @Router /v1/user [get]
// List Users Handler
func ListUsers(c *fiber.Ctx, repo repository.Connection) error {
	ctx, span := tracer.Start(c.UserContext(), "listUserHandler")
	userList, err := usecases.ListUsers(repo, ctx)
	defer span.End()

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(userList)
}

// Get User
// @Summary Get User Service
// @Id get_user
// @version 1.0
// @produce application/json
// @Param id path int true "id"
// @Success 200 {object} db.User
// @Router /v1/user/{id} [get]
// Get User Handler
func GetUser(c *fiber.Ctx, repo repository.Connection) error {
	uid, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	ctx := c.UserContext()
	user, err := usecases.GetUser(repo, ctx, uid)

	if err != nil {
		switch {
		case errors.Is(err, usecases.UserNotFoundError):
			return c.Status(http.StatusNotFound).JSON(err)
		default:
			return c.Status(http.StatusInternalServerError).JSON(err)
		}
	}
	return c.Status(http.StatusOK).JSON(user)
}

// Create User
// @Summary Creates a User
// @Id create_user
// @version 1.0
// @produce application/json
// @Success 200 {object} db.User
// @Param name query string true "name"
// @Router /v1/user [put]
// Create User Handler
func CreateUser(c *fiber.Ctx, repo repository.Connection) error {
	user := &db.User{
		Name: c.Query("name"),
	}
	ctx := c.UserContext()
	userResult, err := usecases.CreateUser(repo, ctx, user)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(userResult)
}

// Update User
// @Summary Update a User
// @Id update_user
// @version 1.0
// @produce application/json
// @Success 200 {object} db.User
// @Param id path string true "id"
// @Param name query string true "name"
// @Router /v1/user/{id} [post]
// Update User Handler
func UpdateUser(c *fiber.Ctx, repo repository.Connection) error {
	ctx := c.UserContext()
	uid, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	user := &db.User{
		Name: c.Query("name"),
		Id:   uid,
	}

	userResult, err := usecases.UpdateUser(repo, ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, usecases.UserNotFoundError):
			return c.Status(http.StatusNotFound).JSON(err)
		default:
			return c.Status(http.StatusInternalServerError).JSON(err)
		}
	}

	return c.Status(http.StatusOK).JSON(userResult)
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
func DeleteUser(c *fiber.Ctx, repo repository.Connection) error {
	uid, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}
	ctx := c.UserContext()

	err = usecases.DeleteUser(repo, ctx, uid)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}
	return c.Status(http.StatusOK).SendString("success")
}
