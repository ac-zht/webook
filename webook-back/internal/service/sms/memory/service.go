package memory

import (
	"context"
	"fmt"
)

type Service struct {
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}

func NewService() *Service {
	return &Service{}
}
