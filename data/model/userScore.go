package model

import uuid "github.com/satori/go.uuid"

//modeling table UserScore
type UserScore struct {
	BaseModel
	UserID uuid.UUID
	User   User

	TestID uuid.UUID
	Test   Test

	TotalNotAnswered   int
	TotalRightAnswered int
	TotalWrongAnswered int
	Score              int
}
