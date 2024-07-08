package saramax

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/ac-zht/webook/pkg/logger"
	"time"
)

type BatchHandler[T any] struct {
	l  logger.Logger
	fn func(msgs []*sarama.ConsumerMessage, t []T) error
}

func (b BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgsCh := claim.Messages()
	const batchSize = 10
	for {
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		done := false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgsCh:
				if !ok {
					cancel()
					return nil
				}
				msgs = append(msgs, msg)
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化消息体失败",
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition),
						logger.Int64("offset", msg.Offset),
						logger.Error(err))
					session.MarkMessage(msg, "")
					continue
				}
				ts = append(ts, t)
			}
		}
		err := b.fn(msgs, ts)
		if err != nil {
			for _, msg := range msgs {
				session.MarkMessage(msg, "")
			}
		} else {

		}
		cancel()
	}
}

func NewBatchHandler[T any](l logger.Logger,
	fn func(msgs []*sarama.ConsumerMessage, t []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{
		l:  l,
		fn: fn,
	}
}
