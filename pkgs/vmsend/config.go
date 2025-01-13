package vmsend

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	VmdbImportUrl string               // VirtualMetrics 导入 URL
	BasicAuth     []string             // 基本认证用户名
	DbFile        string               // 数据库文件路径
	MetricsBucket string               // 存储未发送指标的 Bucket 名称
	MaxRetries    int                  // 最大重试次数
	RetryWaitTime time.Duration        // 重试等待时间
	BatchSize     int                  // 批量发送的指标数量
	PromReg       *prometheus.Registry // Prometheus 注册器
}

func NewConfig(url string, auth ...string) *Config {
	t := &Config{
		DbFile:        "metrics.db",
		MetricsBucket: "UnsentMetrics",
		MaxRetries:    5,
		RetryWaitTime: 2 * time.Second,
		BatchSize:     10,
		VmdbImportUrl: url,
		BasicAuth:     []string{"", ""},
	}
	if len(auth) > 0 {
		t.BasicAuth = t.splitBasicAuth(auth[0])
	}
	return t
}

/*
SetPromReg 设置 Prometheus 注册器
SetPromReg 将提供的 Prometheus 注册器分配给 Config 结构体。
参数:

	reg - Prometheus 注册器实例，用于收集和暴露指标数据。

返回:

	返回 *Config 以支持链式方法调用。
*/
func (t *Config) SetPromReg(reg *prometheus.Registry) *Config {
	t.PromReg = reg
	return t
}

// splitBasicAuth 分割基本认证字符串
func (t *Config) splitBasicAuth(auth string) []string {
	parts := []string{"", ""}
	if len(auth) == 0 {
		return parts
	}
	for i := 0; i < len(auth); i++ {
		if auth[i] == ':' {
			parts[0] = auth[:i]
			parts[1] = auth[i+1:]
			return parts
		}
	}
	return parts
}
