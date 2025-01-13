package examples

import (
	"context"
	"time"

	"github.com/lwmacct/241224-go-template-pkgs/m_log"
	"github.com/lwmacct/241224-go-template-pkgs/m_time"
	"github.com/lwmacct/250113-go-metrics/pkgs/vmsend"
)

type Example struct{}

func (e *Example) mian() {

	// 设置 PromReg 可以定期发送指标
	vmc := vmsend.NewConfig(
		"http://localhost:8428/api/v1/import",
		"user:password",
	).SetPromReg(nil)
	vms, err := vmsend.NewTs(vmc)

	if err != nil {
		m_log.Error(m_log.H{"msg": "vmsend.NewTs", "data": err.Error()})
		return
	}

	m1 := vmsend.NewMetric(
		map[string]string{
			"__name__": "m_250112_test_001",
			"app":      "vmsend",
		},
	)
	m1.AddValue(1.0, time.Now().In(m_time.GetMux().Location).UnixMilli()) // 添加值
	vms.AddMetric(m1)                                                     // 添加指标
	vms.Flush()

	time.Sleep(1 * time.Second)

	// 定期发送指标
	go vms.Ticker(60*time.Second, context.Background())

}
