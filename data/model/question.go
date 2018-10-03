package model

import uuid "github.com/satori/go.uuid"

//modeling table Question
type Question struct {
	BaseModel
	Question        string `gorm:"type:varchar(100);"`
	Answer          string `gorm:"type:varchar(100);"`
	TestID          uuid.UUID
	Test            Test
	QuestionChoices []QuestionChoice
}
