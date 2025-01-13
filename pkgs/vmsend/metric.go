package vmsend

import (
	"encoding/json"
	"sync"

	"github.com/lwmacct/241224-go-template-pkgs/m_to"
)

// 表示单个指标的数据
type Metric struct {
	Metric     map[string]string `json:"metric"`
	Values     []float64         `json:"values"`
	Timestamps []int64           `json:"timestamps"`
	mu         sync.Mutex        `json:"-"`
}

// 初始化一个新的 Metric 实例
func NewMetric(label map[string]string) *Metric {
	return &Metric{
		Metric:     label,
		Values:     make([]float64, 0),
		Timestamps: make([]int64, 0),
	}
}

// 添加一个值和时间戳到指标中
func (m *Metric) AddValue(value float64, timestamp int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Values = append(m.Values, value)
	m.Timestamps = append(m.Timestamps, timestamp)
}

// 添加一个值和时间戳到指标中
func (m *Metric) AddValueAny(value any, timestamp any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Values = append(m.Values, m_to.Float64(value))
	m.Timestamps = append(m.Timestamps, m_to.Int64(timestamp))
}

// 替换指标的值和时间戳
func (m *Metric) SetValues(values []float64, timestamps []int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Values = values
	m.Timestamps = timestamps
}

// 序列化 Metric 为 JSON 字符串
func (m *Metric) ToJSON() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Marshal(m)
}
