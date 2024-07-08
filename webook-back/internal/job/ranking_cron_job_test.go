package job

import (
	"context"
	"errors"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/service"
	svcmocks "github.com/ac-zht/webook/internal/service/mocks"
	"github.com/ac-zht/webook/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestScheduler_Start(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.CronJobService

		wantErr error
		wantJob *TestJob
	}{
		{
			name: "调度了一个任务",
			mock: func(ctrl *gomock.Controller) service.CronJobService {
				svc := svcmocks.NewMockCronJobService(ctrl)
				svc.EXPECT().Preempt(gomock.Any()).
					Return(domain.CronJob{
						Id:         1,
						Name:       "test_job",
						Executor:   "local",
						Cfg:        "hello,world",
						Expression: "my cron expression",
						CancelFunc: func() {},
					}, nil)
				svc.EXPECT().Preempt(gomock.Any()).AnyTimes().
					Return(domain.CronJob{}, errors.New("db 错误"))
				svc.EXPECT().ResetNextTime(gomock.Any(), gomock.Any()).
					Return(errors.New("db 错误"))
				return svc
			},
			wantErr: context.DeadlineExceeded,
			wantJob: &TestJob{
				cnt: 1,
			},
		},
		{
			name: "Executor 没找到",
			mock: func(ctrl *gomock.Controller) service.CronJobService {
				svc := svcmocks.NewMockCronJobService(ctrl)
				svc.EXPECT().Preempt(gomock.Any()).
					Return(domain.CronJob{
						Id:         1,
						Name:       "test_job",
						Executor:   "fake news",
						Cfg:        "hello,world",
						Expression: "my cron expression",
						CancelFunc: func() {},
					}, nil)
				svc.EXPECT().Preempt(gomock.Any()).AnyTimes().
					Return(domain.CronJob{}, errors.New("db 错误"))
				return svc
			},
			wantErr: context.DeadlineExceeded,
			wantJob: &TestJob{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := tc.mock(ctrl)
			scheduler := NewScheduler(svc, logger.NewNoOpLogger())
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			executor := NewLocalFuncExecutor()
			testJob := &TestJob{}
			executor.AddLocalFunc("test_job", testJob.Do)
			scheduler.RegisterExecutor(executor)
			err := scheduler.Start(ctx)
			assert.Error(t, tc.wantErr, err)
			assert.Equal(t, tc.wantJob, testJob)
		})
	}
}

type TestJob struct {
	cnt int
}

func (t *TestJob) Do(ctx context.Context, j domain.CronJob) error {
	t.cnt++
	return nil
}
