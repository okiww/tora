package user

import uuid "github.com/satori/go.uuid"

type result struct {
	ID                 uuid.UUID `json:"id" binding:"required"`
	UserID             uuid.UUID `json:"user_id" binding:"required"`
	Name               string    `json:"name" binding:"required"`
	TotalRightAnswered int       `json:"total_right_answered" binding:"required"`
	TotalWrongAnswered int       `json:"total_wrong_answered" binding:"required"`
	TotalNotAnswered   int       `json:"total_not_answered" binding:"required"`
	Score              int       `json:"score" binding:"required"`
	TimeComplete       string    `json:"time_complete" binding:"required"`
}
