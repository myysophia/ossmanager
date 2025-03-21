package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateObjectKey 生成唯一的对象键名
// ext 是文件扩展名，例如 ".jpg", ".png" 等
func GenerateObjectKey(ext string) string {
	now := time.Now()
	// 生成格式为 年月日/小时分钟秒_随机UUID.扩展名 的对象键
	// 例如: 20230421/143045_550e8400-e29b-41d4-a716-446655440000.jpg
	return fmt.Sprintf("%s/%s_%s%s",
		now.Format("20060102"),
		now.Format("150405"),
		uuid.New().String(),
		ext,
	)
}
