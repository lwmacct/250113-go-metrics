package vmsend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

// Ts 负责管理和发送指标数据
type Ts struct {
	config *Config
	client *resty.Client
	db     *bbolt.DB

	mu  sync.Mutex
	err error
}

// NewTs 初始化一个新的 Ts 实例
func NewTs(config *Config) (*Ts, error) {
	if config == nil {
		return nil, errors.New("配置不能为空")
	}

	t := &Ts{
		config: config,
	}

	// 初始化 Resty 客户端
	t.client = resty.New()

	t.client.DisableWarn = true
	t.client.SetRetryCount(t.config.MaxRetries).
		SetRetryWaitTime(t.config.RetryWaitTime).
		SetRetryMaxWaitTime(t.config.RetryWaitTime * 2).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true // 出现错误时重试
			}
			// 当状态码不为204时重试
			return r.StatusCode() != 204
		})

	// 打开或创建 BoltDB 数据库
	t.db, t.err = bbolt.Open(t.config.DbFile, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if t.err != nil {
		return nil, errors.Wrap(t.err, "无法打开 BoltDB 数据库")
	}

	// 创建存储未发送指标的 Bucket
	t.err = t.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(t.config.MetricsBucket))
		return err
	})
	if t.err != nil {
		t.db.Close()
		return nil, errors.Wrap(t.err, "无法创建或打开 Bucket")
	}
	return t, nil
}

// 将 Metric 持久化到 BoltDB
func (t *Ts) AddMetric(m *Metric) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	//
	data, err := m.ToJSON()
	if err != nil {
		return err
	}
	key := uuid.New().String()
	return t.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(t.config.MetricsBucket))
		return b.Put([]byte(key), data)
	})
}

// 获取所 BoltDB 中所有待发送的 Metrics
func (t *Ts) getPendingMetrics() ([]*Metric, map[string][]byte, error) {
	var metrics []*Metric
	keys := make(map[string][]byte)
	err := t.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(t.config.MetricsBucket))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var metric Metric
			if err := json.Unmarshal(v, &metric); err != nil {
				fmt.Printf("无法解析指标 (key: %s): %v\n", string(k), err)
				return nil // 跳过无法解析的条目
			}
			metrics = append(metrics, &metric)
			keys[string(k)] = v
			return nil
		})
	})
	if err != nil {
		return nil, nil, err
	}
	return metrics, keys, nil
}

func (t *Ts) toJsonLine(metrics []*Metric) ([]byte, error) {
	// 创建一个缓冲区来存储序列化后的 JSON 对象
	var buffer bytes.Buffer
	for _, metric := range metrics {
		metricJSON, err := json.Marshal(metric)
		if err != nil {
			return nil, errors.Wrap(err, "序列化 Metric 失败")
		}
		buffer.Write(metricJSON)
		buffer.WriteByte('\n')
	}
	return buffer.Bytes(), nil
}

// 发送一批 Metrics
func (t *Ts) sendBatch(metrics []*Metric) error {

	body, err := t.toJsonLine(metrics)
	if err != nil {
		return errors.Wrap(err, "序列化 Metrics 失败")
	}

	// 发送 HTTP POST 请求
	resp, err := t.client.R().
		SetDebug(false).
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(t.config.BasicAuth[0], t.config.BasicAuth[1]).
		SetBody(body).
		Post(t.config.VmdbImportUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 204 {
		return fmt.Errorf("收到非204的状态码: %d, 响应体: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// 发送所有待发送的 Metrics
func (t *Ts) Flush() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	metrics, keys, err := t.getPendingMetrics()
	if err != nil {
		return errors.Wrap(err, "获取待发送指标失败")
	}

	if len(metrics) == 0 {
		return nil
	}

	// 发送批量指标
	err = t.sendBatch(metrics)
	if err != nil {
		return errors.Wrap(err, "发送批量指标失败")
	}

	// 发送成功后，删除已发送的指标
	err = t.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(t.config.MetricsBucket))
		for k := range keys {
			if err := b.Delete([]byte(k)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "删除已发送的指标失败")
	}
	return nil
}

// Close 关闭 BoltDB 数据库
func (t *Ts) Close() error {
	return t.db.Close()
}
