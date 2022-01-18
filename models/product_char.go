package models

import "gorm.io/gorm"

type ProductChar struct {
	gorm.Model
	ProductID int
	Name      string
	Value     string
}
