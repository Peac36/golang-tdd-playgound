package models

import "gorm.io/gorm"

type Event struct {
	gorm.Model
	Name   string `validate:"required,min=6"`
	UserID uint
}
