package admin

import (
	"github.com/satori/go.uuid"
)

type testResponse struct {
	ID            uuid.UUID `json:"id" binding:"required"`
	Name          string    `json:"name" binding:"required"`
	Description   string    `json:"description" binding:"required"`
	TotalQuestion int       `json:"total_question" binding:"required"`
}

type testDetailResponse struct {
	ID            uuid.UUID          `json:"id" binding:"required"`
	Name          string             `json:"name" binding:"required"`
	Description   string             `json:"description" binding:"required"`
	TotalQuestion int                `json:"total_question" binding:"required"`
	Questions     []questionResponse `json:"question" binding:"required"`
}

type questionResponse struct {
	ID       uuid.UUID                `json:"id" binding:"required"`
	Question string                   `json:"question" binding:"required"`
	Answer   string                   `json:"answer" binding:"required"`
	Choices  []questionChoiceResponse `json:"choice" binding:"required"`
}

type questionChoiceResponse struct {
	ID     uuid.UUID `json:"id" binding:"required"`
	Key    int       `json:"key" binding:"required"`
	Choice string    `json:"choice" binding:"required"`
}
