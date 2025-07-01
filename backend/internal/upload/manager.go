package upload

import (
	"sync"
	"time"
)

type ChunkInfo struct {
	ChunkNumber int   `json:"chunk_number"`
	ChunkSize   int64 `json:"chunk_size"`
	Uploaded    bool  `json:"uploaded"`
}

type Progress struct {
	Total       int64       `json:"total"`
	Uploaded    int64       `json:"uploaded"`
	Percentage  float64     `json:"percentage"`
	Speed       int64       `json:"speed"` // bytes per second
	StartTime   time.Time   `json:"start_time"`
	UpdateTime  time.Time   `json:"update_time"`
	IsChunked   bool        `json:"is_chunked"`   // 是否为分片上传
	TotalChunks int         `json:"total_chunks"` // 总分片数
	Chunks      []ChunkInfo `json:"chunks"`       // 分片信息
	Status      string      `json:"status"`       // uploading, completed, failed
}

type Manager struct {
	mu          sync.RWMutex
	progresses  map[string]*Progress
	subscribers map[string]map[chan Progress]struct{}
}

func NewManager() *Manager {
	return &Manager{
		progresses:  make(map[string]*Progress),
		subscribers: make(map[string]map[chan Progress]struct{}),
	}
}

var DefaultManager = NewManager()

func (m *Manager) Start(id string, total int64) {
	m.StartWithChunks(id, total, false, 0)
}

func (m *Manager) StartWithChunks(id string, total int64, isChunked bool, totalChunks int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	progress := &Progress{
		Total:       total,
		Uploaded:    0,
		Percentage:  0,
		Speed:       0,
		StartTime:   now,
		UpdateTime:  now,
		IsChunked:   isChunked,
		TotalChunks: totalChunks,
		Status:      "uploading",
	}

	if isChunked && totalChunks > 0 {
		progress.Chunks = make([]ChunkInfo, totalChunks)
		chunkSize := total / int64(totalChunks)
		for i := 0; i < totalChunks; i++ {
			progress.Chunks[i] = ChunkInfo{
				ChunkNumber: i + 1,
				ChunkSize:   chunkSize,
				Uploaded:    false,
			}
		}
		// 调整最后一个分片的大小
		if totalChunks > 0 {
			progress.Chunks[totalChunks-1].ChunkSize = total - chunkSize*int64(totalChunks-1)
		}
	}

	m.progresses[id] = progress
}

func (m *Manager) Update(id string, uploaded int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.progresses[id]; ok {
		if uploaded > p.Uploaded {
			now := time.Now()
			duration := now.Sub(p.UpdateTime).Seconds()
			if duration > 0 {
				p.Speed = int64(float64(uploaded-p.Uploaded) / duration)
			}
			p.Uploaded = uploaded
			p.UpdateTime = now
			if p.Total > 0 {
				p.Percentage = float64(uploaded) / float64(p.Total) * 100
			}
		}

		// 通知订阅者
		for ch := range m.subscribers[id] {
			select {
			case ch <- *p:
			default:
			}
		}
	}
}

func (m *Manager) UpdateChunk(id string, chunkNumber int, uploaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.progresses[id]; ok && p.IsChunked {
		if chunkNumber > 0 && chunkNumber <= len(p.Chunks) {
			p.Chunks[chunkNumber-1].Uploaded = uploaded

			// 重新计算总进度
			var totalUploaded int64
			for _, chunk := range p.Chunks {
				if chunk.Uploaded {
					totalUploaded += chunk.ChunkSize
				}
			}

			now := time.Now()
			duration := now.Sub(p.UpdateTime).Seconds()
			if duration > 0 {
				p.Speed = int64(float64(totalUploaded-p.Uploaded) / duration)
			}

			p.Uploaded = totalUploaded
			p.UpdateTime = now
			if p.Total > 0 {
				p.Percentage = float64(totalUploaded) / float64(p.Total) * 100
			}

			// 通知订阅者
			for ch := range m.subscribers[id] {
				select {
				case ch <- *p:
				default:
				}
			}
		}
	}
}

func (m *Manager) Finish(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.progresses[id]; ok {
		p.Status = "completed"
		p.Percentage = 100
		p.Uploaded = p.Total

		// 最后通知一次订阅者
		for ch := range m.subscribers[id] {
			select {
			case ch <- *p:
			default:
			}
		}
	}

	// 延迟删除进度信息，让客户端有时间接收完成状态
	go func() {
		time.Sleep(5 * time.Second)
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.progresses, id)
		if subs, ok := m.subscribers[id]; ok {
			for ch := range subs {
				close(ch)
			}
			delete(m.subscribers, id)
		}
	}()
}

func (m *Manager) Fail(id string, errorMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.progresses[id]; ok {
		p.Status = "failed"

		// 通知订阅者失败状态
		for ch := range m.subscribers[id] {
			select {
			case ch <- *p:
			default:
			}
		}
	}
}

func (m *Manager) Get(id string) (Progress, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.progresses[id]
	if !ok {
		return Progress{}, false
	}
	return *p, true
}

func (m *Manager) Subscribe(id string) chan Progress {
	ch := make(chan Progress, 10)
	m.mu.Lock()
	if _, ok := m.subscribers[id]; !ok {
		m.subscribers[id] = make(map[chan Progress]struct{})
	}
	m.subscribers[id][ch] = struct{}{}
	if p, ok := m.progresses[id]; ok {
		ch <- *p
	}
	m.mu.Unlock()
	return ch
}

func (m *Manager) Unsubscribe(id string, ch chan Progress) {
	m.mu.Lock()
	if subs, ok := m.subscribers[id]; ok {
		if _, ok := subs[ch]; ok {
			delete(subs, ch)
			close(ch)
		}
		if len(subs) == 0 {
			delete(m.subscribers, id)
		}
	}
	m.mu.Unlock()
}
