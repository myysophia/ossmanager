package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// WebDAVToken WebDAV 访问令牌模型
type WebDAVToken struct {
	Model
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Bucket    string    `gorm:"size:100;not null;index" json:"bucket"`
	TokenHash string    `gorm:"size:255;not null;index" json:"-"` // 不在 JSON 中暴露
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	
	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// SetToken 设置令牌哈希
func (t *WebDAVToken) SetToken(token string) error {
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	t.TokenHash = string(hashedToken)
	return nil
}

// CheckToken 验证令牌
func (t *WebDAVToken) CheckToken(token string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(t.TokenHash), []byte(token))
	return err == nil
}

// IsExpired 检查令牌是否过期
func (t *WebDAVToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid 检查令牌是否有效（未过期且未删除）
func (t *WebDAVToken) IsValid() bool {
	return !t.IsExpired() && t.DeletedAt.Time.IsZero()
}

// TableName 指定表名
func (WebDAVToken) TableName() string {
	return "webdav_tokens"
}

// WebDAVTokenResponse 用于 API 响应的结构体
type WebDAVTokenResponse struct {
	ID        uint      `json:"id"`
	Bucket    string    `json:"bucket"`
	Token     string    `json:"token,omitempty"`     // 只在创建时返回原始令牌
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse 转换为响应结构体
func (t *WebDAVToken) ToResponse() *WebDAVTokenResponse {
	return &WebDAVTokenResponse{
		ID:        t.ID,
		Bucket:    t.Bucket,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
	}
}
