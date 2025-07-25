package models

import "time"

// MD5状态常量
const (
	MD5StatusPending     = "PENDING"     // 待计算
	MD5StatusCalculating = "CALCULATING" // 计算中
	MD5StatusCompleted   = "COMPLETED"   // 已完成
	MD5StatusFailed      = "FAILED"      // 计算失败
)

// OSSFile OSS 文件模型
type OSSFile struct {
	Model
	Filename         string    `gorm:"size:255;not null" json:"filename"`
	OriginalFilename string    `gorm:"size:255;not null" json:"original_filename"`
	FileSize         int64     `gorm:"not null" json:"file_size"`
	MD5              string    `gorm:"size:32" json:"md5"`
	MD5Status        string    `gorm:"size:20;default:'PENDING'" json:"md5_status"` // PENDING, CALCULATING, COMPLETED, FAILED
	StorageType      string    `gorm:"size:20;not null" json:"storage_type"`        // ALIYUN_OSS, AWS_S3, CLOUDFLARE_R2
	Bucket           string    `gorm:"size:100;not null" json:"bucket"`
	ObjectKey        string    `gorm:"size:255;not null" json:"object_key"`
	DownloadURL      string    `gorm:"type:text" json:"download_url,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	UploaderID       uint      `gorm:"not null" json:"uploader_id"`
	Uploader         *User     `json:"uploader,omitempty"`
	UploadIP         string    `gorm:"size:50" json:"upload_ip"`
	Status           string    `gorm:"size:20;default:ACTIVE" json:"status"` // ACTIVE, DELETED
	ConfigID         uint      `gorm:"not null" json:"config_id"`            // 存储配置ID
}

// TableName 指定表名
func (OSSFile) TableName() string {
	return "oss_files"
}
