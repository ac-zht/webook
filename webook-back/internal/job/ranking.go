package job

import (
	"context"
	rlock "github.com/ac-zht/gotools/redis-lock"
	"github.com/ac-zht/webook/internal/service"
	"github.com/ac-zht/webook/pkg/logger"
	"sync"
	"time"
)

//基于redis分布式锁的抢占

type RankingJob struct {
	svc        service.RankingService
	timeout    time.Duration
	lockClient *rlock.Client
	l          logger.Logger
	key        string

	localLock sync.Mutex
	lock      *rlock.Lock
}

func NewRankingJob(
	svc service.RankingService,
	lockClient *rlock.Client,
	l logger.Logger,
	timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc:        svc,
		lockClient: lockClient,
		timeout:    timeout,
		key:        "job:ranking",
		l:          l,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	//不用defer释放本地锁是为保证计算热榜时不阻塞线程释放锁（防止redis已经释放分布式锁时由于本地锁阻塞导致变量无法置为nil，当线程a执行完run方法释放本地锁，
	//同一时刻同实例线程b先于释放分布式锁的线程抢到本地锁，则会认为该实例仍持有分布式锁直接去执行任务，与此同时其他实例可能已经抢占了分布式锁也在执行任务）
	if r.lock == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		var err error
		//拿锁
		r.lock, err = r.lockClient.Lock(ctx, r.key, r.timeout, &rlock.FixedIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      3,
		}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			return nil
		}
		r.localLock.Unlock()
		//续约
		go func() {
			err = r.lock.AutoRefresh(r.timeout/2, r.timeout)
			if err != nil {
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	} else {
		r.localLock.Unlock()
	}
	return r.run()
}

func (r *RankingJob) run() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.RankTopN(ctx)
}

var _ Job = (*RankingJob)(nil)
