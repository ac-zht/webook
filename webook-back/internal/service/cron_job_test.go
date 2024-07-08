package service

import (
	"context"
	"errors"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/repository"
	repomocks "github.com/ac-zht/webook/internal/repository/mocks"
	"github.com/ac-zht/webook/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCronJobService_Preempt(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.CronJobRepository
		wantErr  error
		wantJob  domain.CronJob
		interval time.Duration
	}{
		{
			name: "抢占并续约",
			mock: func(ctrl *gomock.Controller) repository.CronJobRepository {
				repo := repomocks.NewMockCronJobRepository(ctrl)
				repo.EXPECT().Preempt(gomock.Any()).Return(domain.CronJob{
					Id: 1,
				}, nil)
				repo.EXPECT().UpdateUtime(gomock.Any(), int64(1)).Times(3).Return(nil)
				repo.EXPECT().Release(gomock.Any(), int64(1)).Return(nil)
				return repo
			},
			interval: time.Second*3 + time.Millisecond*100,
			wantErr:  nil,
			wantJob: domain.CronJob{
				Id: 1,
			},
		},
		{
			name: "抢占失败",
			mock: func(ctrl *gomock.Controller) repository.CronJobRepository {
				repo := repomocks.NewMockCronJobRepository(ctrl)
				repo.EXPECT().Preempt(gomock.Any()).
					Return(domain.CronJob{}, errors.New("db error"))
				return repo
			},
			interval: time.Second*3 + time.Millisecond*100,
			wantErr:  errors.New("db error"),
			wantJob:  domain.CronJob{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewCronJobService(tc.mock(ctrl), logger.NewNoOpLogger())
			svc.(*cronJobService).refreshInterval = time.Second
			job, err := svc.Preempt(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.NotNil(t, job.CancelFunc)
			cancelFunc := job.CancelFunc
			job.CancelFunc = nil
			assert.Equal(t, tc.wantJob, job)

			time.Sleep(tc.interval)
			cancelFunc()
			time.Sleep(tc.interval)
		})
	}
}
