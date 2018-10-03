package model

//modeling table Test
type Test struct {
	BaseModel
	Name        string `gorm:"type:varchar(100);"`
	Description string `gorm:"type:varchar(255);"`
	Questions   []Question
}
