package mysql

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gorm_prom "gorm.io/plugin/prometheus"
)

type UserRepo struct {
	Db *gorm.DB
}

func New(conn string) (UserRepo, error) {
	userDb, err := gorm.Open(mysql.Open(conn), &gorm.Config{})
	if err != nil {
		return UserRepo{}, err
	}
	return UserRepo{Db: userDb}, err
}

func (r *UserRepo) Initialize(models ...interface{}) {

	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4;",
		"db",
	)

	tx := r.Db.Exec(createSQL)

	r.Db.AutoMigrate(&User{})

	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	r.Db.Use(gorm_prom.New(gorm_prom.Config{
		DBName:          "db", // use `DBName` as metrics label
		RefreshInterval: 15,   // Refresh metrics interval (default 15 seconds)
	}))
}
