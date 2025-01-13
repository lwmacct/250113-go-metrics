// vmsend.pub.go
package vmsend

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

	dataToSend := t.metric
	t.metric = t.metric[:0]
	t.mu.Unlock()

	err := t.sendBatch(dataToSend)
	if err != nil {
		return err
	}

	return nil
}
