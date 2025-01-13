// vmsend.go
package vmsend

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

// Ts 负责管理和发送指标数据
type Ts struct {
	config *Config
	client *resty.Client
	metric [][]byte
	db     *bbolt.DB
	mu     sync.Mutex
	err    error
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
