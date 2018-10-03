package model

import uuid "github.com/satori/go.uuid"

//modeling table UserAnswer
type UserAnswer struct {
	BaseModel
	UserID     uuid.UUID
	TestID     uuid.UUID
	QuestionID uuid.UUID
	ChoiceID   uuid.UUID
	Point      int

	User           User
	Test           Test
	Question       Question
	QuestionChoice QuestionChoice
}
