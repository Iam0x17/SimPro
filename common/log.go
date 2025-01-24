package common

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *zap.Logger

const (
	EventStartService    = "StartService"
	EventStopService     = "StopService"
	EventNewConnection   = "NewConnection"
	EventCloseConnection = "CloseConnection"
	EventExecuteCommand  = "ExecuteCommand"
	EventReplyCommand    = "ReplyCommand"
	EventAccountLogin    = "AccountLogin"
	EventStartShell      = "StartShell"
	EventStopShell       = "StopShell"
)

func InitLogger(verbose bool, logFilePath string) {
	var cores []zapcore.Core

	// 自定义编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.MessageKey = "event"
	encoderConfig.LevelKey = ""

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建控制台写入器
	consoleCore := zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), zap.DebugLevel)
	cores = append(cores, consoleCore)

	if logFilePath != "" {
		// 创建文件写入器
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		fileCore := zapcore.NewCore(encoder, zapcore.AddSync(file), zap.DebugLevel)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)

	var opts []zap.Option
	if verbose {
		opts = append(opts, zap.AddCaller(), zap.AddStacktrace(zap.DPanicLevel))
	}

	Logger = zap.New(core, opts...)
}

func SyncLogger() {
	if Logger != nil {
		err := Logger.Sync()
		if err != nil {
			return
		}
	}
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

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
