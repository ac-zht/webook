package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type CronJob struct {
	Id   int64
	Name string

	Executor   string
	Cfg        string
	Expression string
	//任务下一次的执行时间
	NextTime time.Time

	CancelFunc func()
}

func (j CronJob) Next(t time.Time) time.Time {
	expr := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom |
		cron.Month | cron.Dow |
		cron.Descriptor)
	s, _ := expr.Parse(j.Expression)
	return s.Next(t)
}
