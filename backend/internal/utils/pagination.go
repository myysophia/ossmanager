package utils

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// Pagination 分页参数
type Pagination struct {
	CurrentPage int   `json:"current_page"` // 当前页码
	PageSize    int   `json:"page_size"`    // 每页数量
	Total       int64 `json:"total"`        // 总记录数
	Pages       int   `json:"pages"`        // 总页数
}

// GetPagination 从请求中获取分页参数
func GetPagination(c *gin.Context) *Pagination {
	// 默认分页参数
	page := 1
	pageSize := 10

	// 从查询参数中获取分页参数
	if p := c.Query("page"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil && pageNum > 0 {
			page = pageNum
		}
	}

	if ps := c.Query("limit"); ps != "" {
		if size, err := strconv.Atoi(ps); err == nil && size > 0 {
			pageSize = size
		}
	}

	// 限制每页最大数量为100
	if pageSize > 100 {
		pageSize = 100
	}

	return &Pagination{
		CurrentPage: page,
		PageSize:    pageSize,
	}
}

// Paginate 分页查询
func Paginate(p *Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.CurrentPage - 1) * p.PageSize

		// 克隆一个DB实例用于获取总记录数
		countDB := db
		var count int64
		countDB.Count(&count)

		// 计算总页数
		p.Total = count
		p.Pages = int((count + int64(p.PageSize) - 1) / int64(p.PageSize))

		// 应用分页条件
		return db.Offset(offset).Limit(p.PageSize)
	}
}

// GetPaginationResult 获取分页结果
func GetPaginationResult(p *Pagination, list interface{}) map[string]interface{} {
	return map[string]interface{}{
		"list":         list,
		"current_page": p.CurrentPage,
		"page_size":    p.PageSize,
		"total":        p.Total,
		"pages":        p.Pages,
	}
} 