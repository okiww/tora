package model

import "github.com/jinzhu/gorm"

//modeling table Role
type Role struct {
	gorm.Model
	Name string `gorm:"type:varchar(100);"`
}
