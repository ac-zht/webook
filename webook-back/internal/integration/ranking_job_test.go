package integration

import (
	rlock "github.com/ac-zht/gotools/redis-lock"
	"github.com/ac-zht/webook/internal/integration/startup"
	"github.com/ac-zht/webook/internal/job"
	svcmocks "github.com/ac-zht/webook/internal/service/mocks"
	"github.com/ac-zht/webook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestRankingJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rdb := startup.InitRedis()
	svc := svcmocks.NewMockRankingService(ctrl)
	svc.EXPECT().RankTopN(gomock.Any()).Times(3).Return(nil)
	j := job.NewRankingJob(svc, rlock.NewClient(rdb),
		logger.NewNoOpLogger(), time.Minute)
	c := cron.New(cron.WithSeconds())
	bd := job.NewCronJobBuilder(logger.NewNoOpLogger(),
		prometheus.SummaryOpts{
			Namespace: "go_item",
			Subsystem: "webook",
			Name:      "test",
			Help:      "定时任务测试",
			ConstLabels: map[string]string{
				"instance_id": "my_instance_1",
			},
			Objectives: map[float64]float64{
				0.5:   0.01,
				0.75:  0.01,
				0.90:  0.01,
				0.99:  0.001,
				0.999: 0.0001,
			},
		})
	_, err := c.AddJob("@every 1s", bd.Build(j))
	require.NoError(t, err)
	c.Start()
	time.Sleep(time.Second * 3)
	ctx := c.Stop()
	<-ctx.Done()
}
