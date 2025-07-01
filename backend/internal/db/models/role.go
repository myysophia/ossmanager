package models

// Role 角色模型
type Role struct {
	Model
	Name          string                 `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Description   string                 `gorm:"type:text" json:"description"`
	Users         []*User                `gorm:"many2many:user_roles;" json:"users,omitempty"`
	Permissions   []*Permission          `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	RegionBuckets []*RegionBucketMapping `gorm:"many2many:role_region_bucket_access;" json:"region_buckets,omitempty"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}
