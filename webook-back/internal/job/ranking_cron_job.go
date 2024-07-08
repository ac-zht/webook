package job

import (
	"context"
	"errors"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/service"
	"github.com/ac-zht/webook/pkg/logger"
	"golang.org/x/sync/semaphore"
	"time"
)

//基于mysql分布式任务调度机制的抢占

type CronJob = domain.CronJob

type Executor interface {
	Name() string
	Exec(ctx context.Context, j domain.CronJob) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.CronJob) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, job domain.CronJob) error)}
}

func (l *LocalFuncExecutor) AddLocalFunc(name string,
	fn func(ctx context.Context, j domain.CronJob) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.CronJob) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return errors.New("是不是忘记注册本地方法了？")
	}
	return fn(ctx, j)
}

type Scheduler struct {
	execs     map[string]Executor
	interval  time.Duration
	svc       service.CronJobService
	dbTimeout time.Duration
	l         logger.Logger
	limiter   *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.Logger) *Scheduler {
	return &Scheduler{
		execs:     make(map[string]Executor, 8),
		interval:  time.Second,
		svc:       svc,
		dbTimeout: time.Second,
		l:         l,
		limiter:   semaphore.NewWeighted(100),
	}
}

func (s *Scheduler) RegisterJob(ctx context.Context, j CronJob) error {
	return s.svc.AddJob(ctx, j)
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Start(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			time.Sleep(s.interval)
			continue
		}
		exec, ok := s.execs[j.Executor]
		if !ok {
			s.l.Error("不支持的Executor方式")
			j.CancelFunc()
			continue
		}
		go func() {
			defer func() {
				s.limiter.Release(1)
				j.CancelFunc()
			}()

			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("调度任务执行失败",
					logger.Int64("id", j.Id),
					logger.Error(err1))
				return
			}
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("更新下一次的执行失败", logger.Error(err1))
			}
		}()
	}
}
