package model

//modeling table User
type User struct {
	BaseModel
	Name     string `gorm:"type:varchar(100);"`
	Email    string `gorm:"type:varchar(100);"`
	Password string `gorm:"type:varchar(100);"`
	RoleID   uint
	Role     Role
}
