package models

import (
	"golang.org/x/crypto/bcrypt"
)

// User 用户模型
type User struct {
	Model
	Username string  `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password string  `gorm:"size:255;not null" json:"-"`
	Email    string  `gorm:"size:100;uniqueIndex;not null" json:"email"`
	RealName string  `gorm:"size:100" json:"real_name"`
	Status   bool    `gorm:"default:true" json:"status"`
	Roles    []*Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// SetPassword 设置密码
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
} 