package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	ID         int
	OriginalID string
	Name       string
}
