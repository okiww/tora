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
