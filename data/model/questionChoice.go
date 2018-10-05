package model

import uuid "github.com/satori/go.uuid"

//modeling table QuestionChoice
type QuestionChoice struct {
	BaseModel
	Choice string `gorm:"type:varchar(100);"`
	Key    int    `gorm:"type:varchar(100);"`

	QuestionID uuid.UUID `gorm:"type:char(36)" gorm:"default:18"`
	Question   Question
}
