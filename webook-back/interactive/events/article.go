package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/ac-zht/webook/interactive/repository"
	"github.com/ac-zht/webook/internal/events"
	"github.com/ac-zht/webook/pkg/logger"
	"github.com/ac-zht/webook/pkg/saramax"
	"time"
)

const topicReadEvent = "article_read_event"

type ReadEvent struct {
	Aid int64
	Uid int64
}

var _ events.Consumer = &InteractiveReadEventConsumer{}

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveReadEventConsumer(
	client sarama.Client,
	l logger.Logger,
	repo repository.InteractiveRepository) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		repo:   repo,
		client: client,
		l:      l,
	}
}

func (r *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicReadEvent},
			saramax.NewHandler[ReadEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *InteractiveReadEventConsumer) StartBatch() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicReadEvent},
			saramax.NewBatchHandler[ReadEvent](r.l, r.BatchConsume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage,
	evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := r.repo.IncrReadCnt(ctx, "article", evt.Aid)
	return err
}

func (r *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage,
	evts []ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	bizs := make([]string, 0, len(msgs))
	ids := make([]int64, 0, len(msgs))
	for _, evt := range evts {
		bizs = append(bizs, "article")
		ids = append(ids, evt.Uid)
	}
	return r.repo.BatchIncrReadCnt(ctx, bizs, ids)
}
