package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/myysophia/ossmanager/internal/upload"
)

// UploadProgressHandler 处理上传进度查询和SSE
type UploadProgressHandler struct {
	*BaseHandler
}

func NewUploadProgressHandler() *UploadProgressHandler {
	return &UploadProgressHandler{BaseHandler: NewBaseHandler()}
}

// Init 创建一个新的上传进度任务并返回任务ID
func (h *UploadProgressHandler) Init(c *gin.Context) {
	var req struct {
		Total int64 `json:"total"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	id := uuid.NewString()
	upload.DefaultManager.Start(id, req.Total)
	h.Success(c, gin.H{"id": id})
}

// GetProgress 返回上传进度
func (h *UploadProgressHandler) GetProgress(c *gin.Context) {
	id := c.Param("id")
	if p, ok := upload.DefaultManager.Get(id); ok {
		h.Success(c, p)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": "task not found"})
	}
}

// StreamProgress 使用SSE实时推送进度
func (h *UploadProgressHandler) StreamProgress(c *gin.Context) {
	id := c.Param("id")

	// 检查任务是否存在
	if _, exists := upload.DefaultManager.Get(id); !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// SSE头部设置（中间件已设置，这里再次确保）
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 发送初始连接确认
	c.SSEvent("connected", gin.H{"taskId": id, "timestamp": time.Now().Unix()})
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	// 订阅进度更新
	ch := upload.DefaultManager.Subscribe(id)
	defer upload.DefaultManager.Unsubscribe(id, ch)

	// 心跳定时器，防止连接超时
	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()

	// 客户端断开检测
	notify := c.Writer.CloseNotify()

	for {
		select {
		case p, ok := <-ch:
			if !ok {
				// 通道关闭，发送完成事件
				c.SSEvent("complete", gin.H{"taskId": id})
				if flusher, ok := c.Writer.(http.Flusher); ok {
					flusher.Flush()
				}
				return
			}
			// 发送进度更新
			c.SSEvent("progress", p)
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-heartbeat.C:
			// 发送心跳，保持连接活跃
			c.SSEvent("heartbeat", gin.H{"timestamp": time.Now().Unix()})
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-notify:
			// 客户端断开连接
			return
		}
	}
}
