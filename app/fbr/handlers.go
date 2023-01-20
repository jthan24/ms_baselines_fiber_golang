package fbr

import (
	"errors"
	"fmt"
	"net/http"
	"prom/app/db"
	"prom/app/otel"
	"prom/core/domain/logger"
	"prom/core/domain/repository"
	"prom/core/usecases"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// List Users
// @Summary List Users Service
// @Id list_users
// @version 1.0
// @produce application/json
// @Success 200 {object} []db.User
// @Router /v1/user [get]
// List Users Handler
func ListUsers(c *fiber.Ctx, repo repository.Connection, log logger.Logger) error {
	ctx, span := otel.GetTracerInstance().Start(c.UserContext(), "listUsersHandler")
	userList, err := usecases.ListUsers(repo, ctx)
	defer span.End()

	if err != nil {
		log.Error(ctx, "Error Listing users")
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	log.Info(ctx, "Listed Users")

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
func GetUser(c *fiber.Ctx, repo repository.Connection, log logger.Logger) error {
	uid, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	var inputErrs []*ErrorResponse

	if uid < 1 {
		inputErrs = append(inputErrs, &ErrorResponse{
			FailedField: "id",
			Tag:         "The id must be grater than 1",
			Value:       fmt.Sprint(uid),
		})
		return c.Status(http.StatusBadRequest).JSON(inputErrs)
	}

	ctx, span := otel.GetTracerInstance().Start(c.UserContext(), "GetUserHandler")
	defer span.End()
	user, err := usecases.GetUser(repo, ctx, uid)

	if err != nil {
		switch {
		case errors.Is(err, usecases.UserNotFoundError):
			return c.Status(http.StatusNotFound).JSON(err)
		default:
			log.Error(ctx, "Error Getting user with id", zap.Int("uid", uid))
			return c.Status(http.StatusInternalServerError).JSON(err)
		}
	}

	log.Info(ctx, "Got User with id", zap.Int("uid", uid))
	return c.Status(http.StatusOK).JSON(user)
}

// Create User
// @Summary Creates a User
// @Id create_user
// @version 1.0
// @produce application/json
// @Success 200 {object} db.User
// @Param name query string true "name"
// @Router /v1/user [post]
// Create User Handler
func CreateUser(c *fiber.Ctx, repo repository.Connection, log logger.Logger) error {
	name := c.Query("name")
	user := &db.User{
		Name: name,
	}

	errors := ValidateStruct(*user)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}
	ctx, span := otel.GetTracerInstance().Start(c.UserContext(), "CreateUserHandler")
	defer span.End()
	userResult, err := usecases.CreateUser(repo, ctx, user)

	if err != nil {
		log.Error(ctx, "Error creating user with id", zap.String("user-name", name))
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	log.Info(ctx, "Created user with name", zap.String("user-name", name))
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
// @Router /v1/user/{id} [put]
// Update User Handler
func UpdateUser(c *fiber.Ctx, repo repository.Connection, log logger.Logger) error {
	uid, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	user := &db.User{
		Name: c.Query("name"),
		Id:   uid,
	}

	ctx, span := otel.GetTracerInstance().Start(c.UserContext(), "UpdateUserHandler")
	defer span.End()
	userResult, err := usecases.UpdateUser(repo, ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, usecases.UserNotFoundError):
			return c.Status(http.StatusNotFound).JSON(err)
		default:
		  log.Error(ctx, "Error updating user with id", zap.Int("uid", uid))
			return c.Status(http.StatusInternalServerError).JSON(err)
		}
	}

	log.Info(ctx, "Updated user with id", zap.Int("uid", uid))

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
func DeleteUser(c *fiber.Ctx, repo repository.Connection, log logger.Logger) error {
	uid, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}
	ctx, span := otel.GetTracerInstance().Start(c.UserContext(), "DeleteUserHandler")
	defer span.End()

	err = usecases.DeleteUser(repo, ctx, uid)
	if err != nil {
	  log.Error(ctx, "Error deleting user with id", zap.Int("uid", uid))
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	log.Info(ctx, "Deleted user with id", zap.Int("uid", uid))
	return c.Status(http.StatusOK).SendString("success")
}
