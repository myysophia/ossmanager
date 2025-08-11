package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateObjectKey 生成唯一的对象键名，包含用户名目录
// ext 是文件扩展名，例如 ".jpg", ".png" 等
// username 是当前用户名
func GenerateObjectKey(username, ext string) string {
	now := time.Now()
	// 生成格式为 用户名/年月日/小时分钟秒_随机UUID.扩展名 的对象键
	// 例如: alice/20230421/143045_550e8400-e29b-41d4-a716-446655440000.jpg
	return fmt.Sprintf("%s/%s/%s_%s%s",
		username,
		now.Format("20060102"),
		now.Format("150405"),
		uuid.New().String(),
		ext,
	)
}

// GenerateFixedObjectKey 生成固定的对象键名，基于用户名和原始文件名
// originalFilename 是用户上传的原始文件名
// username 是当前用户名
func GenerateFixedObjectKey(username, originalFilename string) string {
	// 生成格式为 用户名/原始文件名 的固定对象键
	// 例如: alice/document.pdf
	return fmt.Sprintf("%s/%s", username, originalFilename)
}

// GenerateFixedObjectKeyWithPath 生成包含自定义路径的固定对象键名
// username 是当前用户名
// customPath 是用户指定的自定义路径，已经过清理
// originalFilename 是用户上传的原始文件名
func GenerateFixedObjectKeyWithPath(username, customPath, originalFilename string) string {
	// 生成格式为 用户名/自定义路径/原始文件名 的固定对象键
	// 例如: alice/文档/图片/2024/document.pdf
	return fmt.Sprintf("%s/%s/%s", username, customPath, originalFilename)
}
