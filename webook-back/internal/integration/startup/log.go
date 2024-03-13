package startup

import "github.com/zht-account/webook/pkg/logger"

func InitLog() logger.Logger {
	return logger.NewNoOpLogger()
}
