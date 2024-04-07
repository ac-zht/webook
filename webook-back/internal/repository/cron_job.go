package repository

import (
	"context"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository/dao"
	"time"
)

var ErrNoMoreJob = dao.ErrNoMoreJob

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.CronJob, error)
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
	UpdateUtime(ctx context.Context, id int64) error
	Release(ctx context.Context, id int64) error
	AddJob(ctx context.Context, j domain.CronJob) error
}

type PreemptCronJobRepository struct {
	dao dao.JobDAO
}

func (p *PreemptCronJobRepository) Preempt(ctx context.Context) (domain.CronJob, error) {
	j, err := p.dao.Preempt(ctx)
	if err != nil {
		return domain.CronJob{}, err
	}
	return p.toDomain(j), nil
}

func (p *PreemptCronJobRepository) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	return p.dao.UpdateNextTime(ctx, id, t)
}

func (p *PreemptCronJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return p.dao.UpdateUtime(ctx, id)
}

func (p *PreemptCronJobRepository) Release(ctx context.Context, id int64) error {
	return p.dao.Release(ctx, id)
}

func (p *PreemptCronJobRepository) AddJob(ctx context.Context, j domain.CronJob) error {
	return p.dao.Insert(ctx, p.toEntity(j))
}

func (p *PreemptCronJobRepository) toEntity(j domain.CronJob) dao.Job {
	return dao.Job{
		Id:         j.Id,
		Name:       j.Name,
		Executor:   j.Executor,
		Cfg:        j.Cfg,
		Expression: j.Expression,
		NextTime:   j.NextTime.UnixMilli(),
	}
}

func (p *PreemptCronJobRepository) toDomain(j dao.Job) domain.CronJob {
	return domain.CronJob{
		Id:         j.Id,
		Name:       j.Name,
		Expression: j.Expression,
		Cfg:        j.Cfg,
		Executor:   j.Executor,
		NextTime:   time.UnixMilli(j.NextTime),
	}
}
