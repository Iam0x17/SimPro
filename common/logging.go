package common

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// setupServiceLogger 为每个服务创建独立的日志记录器
func SetupServiceLogger(serviceName string, logToFile bool) *logrus.Logger {
	logger := logrus.New()
	if logToFile {
		logFile, err := os.OpenFile(fmt.Sprintf("%s.log", serviceName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logrus.Fatalf("打开 %s 日志文件失败: %v", serviceName, err)
		}
		logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
	} else {
		logger.SetOutput(os.Stdout)
	}

	// 设置日志格式，带上服务名称字段
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "时间",
			logrus.FieldKeyLevel: "级别",
			logrus.FieldKeyMsg:   serviceName,
		},
	})

	return logger
}
