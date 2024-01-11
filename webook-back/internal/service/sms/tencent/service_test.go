package tencent

import (
	"context"
	"github.com/magiconair/properties/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestSender(t *testing.T) {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		t.Fatal()
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")

	c, err := sms.NewClient(common.NewCredential(secretId, secretKey), "ap-guangzhou", profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(c, "", "")

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:    "发送验证码",
			tplId:   "",
			params:  []string{},
			numbers: []string{},
		},
	}
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
