// vmsend.pub.go
package vmsend

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

// 将 Metric 持久化到 BoltDB
func (t *Ts) AddMetric(m *Metric) error {
	t.mu.Lock()
	defer t.mu.Unlock()

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
