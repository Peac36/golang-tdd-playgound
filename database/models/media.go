package models

import "gorm.io/gorm"

type Media struct {
	gorm.Model
	Name     string
	Size     int
	Order    int
	Provider string
	Path     string
	EventId  uint
}
