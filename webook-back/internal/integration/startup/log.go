package startup

import "github.com/ac-zht/webook/pkg/logger"

func InitLog() logger.Logger {
	return logger.NewNoOpLogger()
}
