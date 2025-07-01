package models

// Permission 权限模型
type Permission struct {
	Model
	Name        string  `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	Resource    string  `gorm:"size:100;not null" json:"resource"`
	Action      string  `gorm:"size:50;not null" json:"action"`
	Roles       []*Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
} 