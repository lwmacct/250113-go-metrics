// vmsend.priv.go
package vmsend

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/lwmacct/241224-go-template-pkgs/m_log"
	// "go.etcd.io/bbolt"
)

// 发送一批 Metrics
func (t *Ts) sendBatch(data [][]byte) error {
	var buffer bytes.Buffer
	for _, metric := range data {
		metricJSON, err := json.Marshal(metric)
		if err != nil {
			m_log.Info(m_log.H{"err": err.Error()})
			continue
		}
		buffer.Write(metricJSON)
		buffer.WriteByte('\n')
	}

	resp, err := t.client.R().
		SetDebug(false).
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(t.config.BasicAuth[0], t.config.BasicAuth[1]).
		SetBody(buffer.Bytes()).
		Post(t.config.VmdbImportUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 204 {
		return fmt.Errorf("收到非204的状态码: %d, 响应体: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
