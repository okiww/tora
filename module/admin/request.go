package admin

type testRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description" binding:"required"`
	TotalQuestion int    `json:"total_question" binding:"required"`
}

type questionRequest struct {
	TestID    string `json:"test_id" binding:"required"`
	Questions []questions
}

type questions struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
	Choices  []choices
}

type choices struct {
	Choice string `json:"choice" binding:"required"`
}

type updateTestRequest struct {
	TestID        string `json:"test_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description" binding:"required"`
	TotalQuestion int    `json:"total_question" binding:"required"`
}

type updateQuestionRequest struct {
	QuestionID string `json:"question_id" binding:"required"`
	Question   string `json:"question" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

type updateQuestionChoiceRequest struct {
	ChoiceID string `json:"choice_id" binding:"required"`
	Choice   string `json:"choice" binding:"required"`
}

type deleteTestRequest struct {
	TestID string `json:"test_id" binding:"required"`
}

type deleteQuestionRequest struct {
	QuestionID string `json:"question_id" binding:"required"`
}

type deleteChoiceRequest struct {
	QuestionID string `json:"question_id" binding:"required"`
	ChoiceID   string `json:"choice_id" binding:"required"`
}
