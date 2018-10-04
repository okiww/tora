package model

//modeling table Test
type Test struct {
	BaseModel
	Name          string `json:"name" gorm:"type:varchar(100);"`
	Description   string `json:"description" gorm:"type:varchar(255);"`
	TotalQuestion int    `json:"total_question"`

	Questions []Question `json:"questions"`
}
