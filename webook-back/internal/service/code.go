package service

import (
	"context"
	"github.com/ac-zht/webook/internal/repository"
	"github.com/ac-zht/webook/internal/service/sms"
	"math/rand"
	"strconv"
)

const codeTplId = ""
const codeLength = 6

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type SMSCodeService struct {
	sms  sms.Service
	repo repository.CodeRepository
}

func (s *SMSCodeService) Send(ctx context.Context, biz string, phone string) error {
	code := s.generate()
	err := s.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	return s.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (s *SMSCodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	ok, err := s.repo.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooManyTimes {
		return false, nil
	}
	return ok, err
}

func (s *SMSCodeService) generate() string {
	var code string
	for i := 0; i < codeLength; i++ {
		code += strconv.Itoa(rand.Intn(9))
	}
	return code
}

func NewSMSCodeService(svc sms.Service, repo repository.CodeRepository) CodeService {
	return &SMSCodeService{
		sms:  svc,
		repo: repo,
	}
}
