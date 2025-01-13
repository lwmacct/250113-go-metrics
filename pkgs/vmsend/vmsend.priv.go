// vmsend.priv.go
package vmsend

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

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
