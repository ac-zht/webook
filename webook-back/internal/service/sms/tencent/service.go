package tencent

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.TemplateParamSet = common.StringPtrs(args)
	req.PhoneNumberSet = common.StringPtrs(numbers)
	req.SetContext(ctx)
	resp, err := s.client.SendSms(req)
	zap.L().Debug("调用腾讯短信服务",
		zap.Any("req", req),
		zap.Any("resp", resp))
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送失败，code: %s, 原因：%s", *status.Code, *status.Message)
		}
	}
	return nil
}

func NewService(c *sms.Client, appId string, signName string) *Service {
	return &Service{
		client:   c,
		appId:    ekit.ToPtr[string](appId),
		signName: ekit.ToPtr[string](signName),
	}
}
