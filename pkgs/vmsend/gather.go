package vmsend

import (
	"context"
	"time"

	"github.com/lwmacct/241224-go-template-pkgs/m_log"
	"github.com/lwmacct/241224-go-template-pkgs/m_time"
	"github.com/pkg/errors"
)

func (t *Ts) Ticker(d time.Duration, ctx context.Context) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	fc := func() {
		if err := t.Gather(); err != nil {
			m_log.Error(m_log.H{"msg": "vmsendObj.Gather", "data": err.Error()})
		}
	}
	fc()
	for {
		select {
		case <-ticker.C:
			fc()
		case <-ctx.Done():
			return
		}
	}
}

func (t *Ts) Gather() error {

	if t.config.PromReg == nil {
		return errors.New("prometheus registry is nil")
	}
	metricFamilies, err := t.config.PromReg.Gather()
	if err != nil {
		return errors.Wrap(err, "prometheus registry gather error")
	}

	UnixMilli := time.Now().In(m_time.GetMux().Location).UnixMilli()
	// 遍历所有指标族
	for _, mf := range metricFamilies {
		for _, m := range mf.GetMetric() {
			metricName := mf.GetName()
			if m.Gauge != nil {
				mapp := map[string]string{
					"__name__": metricName,
				}
				rawLabel := m.GetLabel()
				for _, l := range rawLabel {
					mapp[l.GetName()] = l.GetValue()
				}
				md := NewMetric(mapp)
				// 毫秒时间戳
				md.AddValue(m.Gauge.GetValue(), UnixMilli)
				t.AddMetric(md)
			}
		}
	}

	if err := t.Flush(); err != nil {
		return errors.Wrap(err, "flush error")
	}
	return nil
}
