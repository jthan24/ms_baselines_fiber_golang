package usecases

import (
	"context"
	"errors"
	"fmt"
	"prom/app/db"
	"prom/app/otel"
	"prom/core/domain/repository"

	"gorm.io/gorm"
)

var (
	UserNotFoundError = errors.New("User Not found")
)

func ListUsers(conn repository.Connection, parentCtx context.Context) ([]*db.User, error) {
	ctx, span := otel.GetTracerInstance().Start(parentCtx, "listUsersUC")
	defer span.End()
	userList := make([]*db.User, 0)
	tx := conn.WithContext(ctx).Find(&userList)

	if tx.Error != nil {
    err := fmt.Errorf("Cannot get users in listUsersUC: %w", tx.Error)
    span.RecordError(err)
		return nil, err
	}

	return userList, nil
}

func GetUser(conn repository.Connection, parentCtx context.Context, uid int) (*db.User, error) {
	ctx, span := otel.GetTracerInstance().Start(parentCtx, "getUserUC")
	defer span.End()

	user := &db.User{}
	tx := conn.WithContext(ctx).Where("id = ?", uid).Find(user)

	if tx.RowsAffected == 0 {
		return nil, UserNotFoundError
	}

	if tx.Error != nil {
    err := fmt.Errorf("Cannot get user with id %d in getUsersUC: %w", uid, tx.Error)
    span.RecordError(err)
		return nil, err
	}
	return user, nil
}

func CreateUser(
	conn repository.Connection,
	parentCtx context.Context,
	user *db.User,
) (*db.User, error) {
	ctx, span := otel.GetTracerInstance().Start(parentCtx, "createUserUC")
	defer span.End()

	tx := conn.WithContext(ctx).Create(user)

	if tx.Error != nil {
    err := fmt.Errorf("Cannot create user in createUsersUC: %w", tx.Error)
		span.RecordError(err)
		return nil, err
	}
	return user, nil
}

func UpdateUser(
	conn repository.Connection,
	parentCtx context.Context,
	user *db.User,
) (*db.User, error) {
	ctx, span := otel.GetTracerInstance().Start(parentCtx, "updateUserUC")
	defer span.End()

	tx := conn.WithContext(ctx).Where("id = ?", user.Id).Updates(user)

	// Good enough if an extra read is not acceptable
	if tx.RowsAffected == 0 {
		return nil, UserNotFoundError
	}

	if tx.Error != nil {
		switch {
		case errors.Is(tx.Error, gorm.ErrRecordNotFound):
			return nil, UserNotFoundError
		default:
      err := fmt.Errorf("Cannot create user in updateUsersUC: %w", tx.Error)
		  span.RecordError(err)
			return nil, err
		}
	}
	return user, nil
}

func DeleteUser(conn repository.Connection, parentCtx context.Context, uid int) error {
	ctx, span := otel.GetTracerInstance().Start(parentCtx, "createUserUC")
	defer span.End()

	tx := conn.WithContext(ctx).Delete(&db.User{
		Id: uid,
	})

	if tx.Error != nil {
    err := fmt.Errorf("Cannot delete user %d in deleteUsersUC: %w", uid, tx.Error)
		span.RecordError(err)
		return err
	}
	return nil
}
