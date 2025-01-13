// vmsend.pub.go
package vmsend

import (
	"bytes"

	"github.com/pkg/errors"
)

func (t *Ts) AddMetric(data []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.metric = append(t.metric, data)
}

func (t *Ts) Flush() error {
	t.mu.Lock()
	if len(t.metric) == 0 {
		t.mu.Unlock()
		return nil
	}

	var buffer bytes.Buffer
	for _, metric := range t.metric {
		buffer.Write(metric)
		buffer.WriteByte('\n')
	}
	t.metric = t.metric[:0]
	t.mu.Unlock()

	resp, err := t.client.R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(t.config.BasicAuth[0], t.config.BasicAuth[1]).
		SetBody(buffer.Bytes()).
		Post(t.config.VmdbImportUrl)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 204 {
		return errors.New("http status code not 204")
	}
	return nil
}
