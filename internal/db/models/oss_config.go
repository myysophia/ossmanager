package models

// OSSConfig OSS 配置模型
type OSSConfig struct {
	Model
	Name          string `gorm:"size:100;not null" json:"name"`
	StorageType   string `gorm:"size:20;not null" json:"storage_type"` // ALIYUN_OSS, AWS_S3, CLOUDFLARE_R2
	AccessKey     string `gorm:"size:255;not null" json:"-"`
	SecretKey     string `gorm:"size:255;not null" json:"-"`
	Endpoint      string `gorm:"size:255;not null" json:"endpoint"`
	Bucket        string `gorm:"size:100;not null" json:"bucket"`
	Region        string `gorm:"size:50" json:"region"`
	IsDefault     bool   `gorm:"default:false" json:"is_default"`
	URLExpireTime int    `gorm:"default:86400" json:"url_expire_time"` // URL过期时间（秒），默认24小时
}

// TableName 指定表名
func (OSSConfig) TableName() string {
	return "oss_configs"
}
