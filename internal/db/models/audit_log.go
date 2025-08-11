package models

// AuditLog 审计日志模型
type AuditLog struct {
	Model
	UserID       uint   `gorm:"index" json:"user_id"`
	Username     string `gorm:"size:50" json:"username"`
	Action       string `gorm:"size:50;not null" json:"action"` // LOGIN, UPLOAD, DOWNLOAD, DELETE, etc.
	ResourceType string `gorm:"size:50" json:"resource_type"`   // FILE, USER, ROLE, etc.
	ResourceID   string `gorm:"size:100" json:"resource_id"`
	Details      string `gorm:"type:jsonb" json:"details"`
	IPAddress    string `gorm:"size:50" json:"ip_address"`
	UserAgent    string `gorm:"type:text" json:"user_agent"`
	Status       string `gorm:"size:20;not null" json:"status"` // SUCCESS, FAILED
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}
