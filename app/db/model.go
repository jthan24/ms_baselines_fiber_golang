package db

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	CreatedAt time.Time      `yaml:"-" json:"-"`
	UpdatedAt time.Time      `yaml:"-" json:"-"`
	DeletedAt gorm.DeletedAt `yaml:"-" json:"-" gorm:"index"`
}

type User struct {
	Base
	Id   int    `yaml:"id"   json:"id"   gorm:"primaryKey"`
	Name string `yaml:"name" json:"name" validate:"required,min=10,max=50"`
}
