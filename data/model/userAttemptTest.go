package model

import uuid "github.com/satori/go.uuid"

//modeling table UserAttempTask
type UserAttemptTest struct {
	BaseModel
	UserID uuid.UUID
	User   User

	TestID uuid.UUID
	Test   Test
}
