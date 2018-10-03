package model

import uuid "github.com/satori/go.uuid"

//modeling table QuestionChoice
type QuestionChoice struct {
	BaseModel
	QuestionOption string `gorm:"type:varchar(100);"`
	key            string `gorm:"type:varchar(100);"`

	QuestionID uuid.UUID
	Question   Question
}
