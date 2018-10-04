package model

import uuid "github.com/satori/go.uuid"

//modeling table UserAnswer
type UserAnswer struct {
	BaseModel
	UserID     uuid.UUID `gorm:"type:char(36)" gorm:"default:18"`
	TestID     uuid.UUID `gorm:"type:char(36)" gorm:"default:18"`
	QuestionID uuid.UUID `gorm:"type:char(36)" gorm:"default:18"`
	ChoiceID   uuid.UUID `gorm:"type:char(36)" gorm:"default:18"`
	Point      int

	User           User
	Test           Test
	Question       Question
	QuestionChoice QuestionChoice
}
