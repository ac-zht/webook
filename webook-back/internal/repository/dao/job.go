package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

var ErrNoMoreJob = gorm.ErrRecordNotFound

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
	UpdateUtime(ctx context.Context, id int64) error
	Release(ctx context.Context, id int64) error
	Insert(ctx context.Context, j Job) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func (dao *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		now := time.Now().UnixMilli()
		var j Job
		err := db.Where(
			"next_time <= ? AND status = ?",
			now, jobStatusWaiting).First(&j).Error
		if err != nil {
			return Job{}, err
		}
		res := db.Model(&Job{}).
			Where("id = ? AND version=?", j.Id, j.Version).
			Updates(map[string]any{
				"utime":   now,
				"version": j.Version + 1,
				"status":  jobStatusRunning,
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 1 {
			return j, nil
		}
	}
}

func (dao *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id=?", id).Updates(map[string]any{
		"utime":     time.Now().UnixMilli(),
		"next_time": t.UnixMilli(),
	}).Error
}

func (dao *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id=?", id).Updates(map[string]any{
		"utime": time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMJobDAO) Release(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMJobDAO) Insert(ctx context.Context, j Job) error {
	now := time.Now().UnixMilli()
	j.Ctime = now
	j.Utime = now
	return dao.db.WithContext(ctx).Create(&j).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(256);unique"`
	Executor   string
	Cfg        string
	Expression string
	Version    int64
	//可建next_time和status的联合索引
	NextTime int64 `gorm:"index:status_next_index"`
	Status   int   `gorm:"index:status_next_index"`
	Ctime    int64
	Utime    int64
}

const (
	jobStatusWaiting = iota
	jobStatusRunning
)
