package models

// RegionBucketMapping 地域-桶映射模型
type RegionBucketMapping struct {
	Model
	RegionCode string  `gorm:"size:50;not null;index" json:"region_code"`  // 地域代码 (e.g., 'us-east-1', 'cn-north-1')
	BucketName string  `gorm:"size:255;not null;index" json:"bucket_name"` // 桶的名称
	Roles      []*Role `gorm:"many2many:role_region_bucket_access;" json:"roles,omitempty"`
}

// TableName 指定表名
func (RegionBucketMapping) TableName() string {
	return "region_bucket_mapping"
}
