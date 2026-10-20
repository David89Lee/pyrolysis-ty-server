package sse

import "sync"

// SampleEvent 一次采样结果，作为 SSE 事件负载推送给前端
type SampleEvent struct {
	Timestamp      string  `json:"timestamp"`
	Temperature    float64 `json:"temperature"`     // °C
	Humidity       float64 `json:"humidity"`        // %
	PM1_0          uint16  `json:"pm1_0"`           // μg/m³
	PM2_5          uint16  `json:"pm2_5"`           // μg/m³
	PM10           uint16  `json:"pm10"`            // μg/m³
	TVOCSample     uint16  `json:"tvoc_sample"`     // ppb
	TVOCBackground uint16  `json:"tvoc_background"` // ppb
	Pyrolysis      uint16  `json:"pyrolysis"`       // ppb
	StatusCode     uint16  `json:"status_code"`
	SensorNormal   bool    `json:"sensor_normal"` // 传感器是否正常
	Alarming       bool    `json:"alarming"`      // 是否报警
	Error          string  `json:"error,omitempty"`
}

// Hub 采样数据广播中心：采集端 Publish，HTTP 层通过 Subscribe 获取事件通道
type Hub struct {
	mu      sync.RWMutex
	clients map[chan SampleEvent]struct{}
	bufSize int
}

// DefaultHub 全局默认 Hub（采集进程只有一个）
var DefaultHub = NewHub(16)

// NewHub 创建 Hub，bufSize 为每个客户端的事件缓冲大小
func NewHub(bufSize int) *Hub {
	if bufSize <= 0 {
		bufSize = 16
	}
	return &Hub{
		clients: make(map[chan SampleEvent]struct{}),
		bufSize: bufSize,
	}
}

// Subscribe 订阅事件，返回的 cancel 用于取消订阅（可重复调用、幂等）
func (h *Hub) Subscribe() (<-chan SampleEvent, func()) {
	ch := make(chan SampleEvent, h.bufSize)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.clients, ch)
			h.mu.Unlock()
			close(ch)
		})
	}
	return ch, cancel
}

// Publish 广播一条事件；客户端消费过慢（缓冲已满）时丢弃该条，避免阻塞采集协程
func (h *Hub) Publish(ev SampleEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- ev:
		default:
		}
	}
}

// ClientCount 返回当前订阅的客户端数量
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
