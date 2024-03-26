package metric

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zht-account/webook/internal/service/sms"
	"time"
)

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func (p PrometheusDecorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		p.vector.WithLabelValues(tplId).
			Observe(float64(duration.Milliseconds()))
	}()
	return p.svc.Send(ctx, tplId, args, numbers...)
}

func NewPrometheusDecorator(svc sms.Service,
	namespace string,
	subsystem string,
	instanceId string,
	name string) *PrometheusDecorator {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		ConstLabels: map[string]string{
			"instance_id": instanceId,
		},
		Objectives: map[float64]float64{
			0.9:   0.01,
			0.95:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"tpl"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		vector: vector,
		svc:    svc,
	}
}
