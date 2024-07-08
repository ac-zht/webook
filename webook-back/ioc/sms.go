package ioc

import (
	"github.com/ac-zht/webook/internal/service/sms"
	"github.com/ac-zht/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
	//credential := common.NewCredential(
	//	os.Getenv("TENCENTCLOUD_SECRET_ID"),
	//	os.Getenv("TENCENTCLOUD_SECRET_KEY"),
	//)
	//smsClient, _ := _sms.NewClient(credential, "ap-guangzhou", profile.NewClientProfile())
	//return tencent.NewService(smsClient, "", "")
}
