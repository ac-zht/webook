package ioc

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	_sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/zht-account/webook/internal/service/sms"
	"github.com/zht-account/webook/internal/service/sms/tencent"
	"os"
)

func InitSMSService() sms.Service {
	credential := common.NewCredential(
		os.Getenv("TENCENTCLOUD_SECRET_ID"),
		os.Getenv("TENCENTCLOUD_SECRET_KEY"),
	)
	smsClient, _ := _sms.NewClient(credential, "ap-guangzhou", profile.NewClientProfile())
	return tencent.NewService(smsClient, "", "")
}
