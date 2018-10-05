package user

type answerRequest struct {
	TestID  string       `json:"test_id" binding:"required"`
	Answers []answerData `json:"answers" binding:"required"`
}

type answerData struct {
	QuestionID string `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

type attempRequest struct {
	TestID string `json:"test_id" binding:"required"`
}
