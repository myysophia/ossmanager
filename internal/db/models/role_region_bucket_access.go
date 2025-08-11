package models

// RoleRegionBucketAccess 角色-地域桶访问权限模型
type RoleRegionBucketAccess struct {
	Model
	RoleID                uint                 `gorm:"not null;index" json:"role_id"`
	Role                  *Role                `json:"role,omitempty"`
	RegionBucketMappingID uint                 `gorm:"not null;index" json:"region_bucket_mapping_id"`
	RegionBucketMapping   *RegionBucketMapping `json:"region_bucket_mapping,omitempty"`
}

// TableName 指定表名
func (RoleRegionBucketAccess) TableName() string {
	return "role_region_bucket_access"
}
