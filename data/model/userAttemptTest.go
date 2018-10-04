package model

import uuid "github.com/satori/go.uuid"

//modeling table UserAttempTask
type UserAttemptTest struct {
	BaseModel
	UserID uuid.UUID `gorm:"type:char(36)" gorm:"default:18"`
	User   User

	TestID uuid.UUID `gorm:"type:char(36)" gorm:"default:18"`
	Test   Test
}
