package vmsend

import (
	"encoding/json"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/lwmacct/241224-go-template-pkgs/pkgs/m_to"
)

// 表示单个指标的数据
type Metrics struct {
	Metric     map[string]string `json:"metric"`
	Values     []float64         `json:"values"`
	Timestamps []int64           `json:"timestamps"`

	lock bool `json:"-"`
	mu   sync.Mutex
}

// 初始化一个新的 Metric 实例
func NewMetrics(label map[string]string) *Metrics {
	return &Metrics{
		Metric:     label,
		Values:     make([]float64, 0),
		Timestamps: make([]int64, 0),
		lock:       false,
	}
}

// 添加一个值和时间戳到指标中
func (m *Metrics) AddValue(value float64, timestamp int64) {
	if m.lock {
		m.mu.Lock()
		defer m.mu.Unlock()
	}

	m.Values = append(m.Values, value)
	m.Timestamps = append(m.Timestamps, timestamp)
}

// 添加一个值和时间戳到指标中
func (m *Metrics) AddValueAny(value any, timestamp any) {
	if m.lock {
		m.mu.Lock()
		defer m.mu.Unlock()
	}
	m.Values = append(m.Values, m_to.Float64(value))
	m.Timestamps = append(m.Timestamps, m_to.Int64(timestamp))
}

// 设置是否加锁
func (m *Metrics) SetLock(lock bool) {
	m.lock = lock
}

// 添加一个标签到指标中
func (m *Metrics) SetLabel(key string, value string) {
	if m.lock {
		m.mu.Lock()
		defer m.mu.Unlock()
	}
	m.Metric[key] = value
}

// 替换所有标签
func (m *Metrics) SetLabels(label map[string]string) {
	if m.lock {
		m.mu.Lock()
		defer m.mu.Unlock()
	}
	m.Metric = label
}

// 替换指标的值和时间戳
func (m *Metrics) SetValues(values []float64, timestamps []int64) {
	if m.lock {
		m.mu.Lock()
		defer m.mu.Unlock()
	}
	m.Values = values
	m.Timestamps = timestamps
}

// 序列化 Metric 为 JSON 字符串
func (m *Metrics) ToJSON() []byte {
	if m.lock {
		m.mu.Lock()
		defer m.mu.Unlock()
	}
	data, _ := json.Marshal(m)
	return data
}

// 发送 Metrics 到指定的 URL
func (m *Metrics) Send(client *resty.Client, url string) (*resty.Response, error) {
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(m.ToJSON()).
		Post(url)
	return resp, err
}
