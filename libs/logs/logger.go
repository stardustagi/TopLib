package logs

import (
	"encoding/json"
	"os"

	"github.com/stardustagi/TopLib/utils"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func Init(logConfigJson []byte, level zapcore.Level) {
	// * lumberjack.Logger 用于日志轮转
	var logConfig *lumberjack.Logger
	var err error
	if logConfig, err = utils.Bytes2Struct[*lumberjack.Logger](logConfigJson); err != nil {
		panic("Failed to parse log configuration: " + err.Error())
	}
	// 日志级别
	var encoderCfg zapcore.EncoderConfig
	// level := zapcore.InfoLevel
	if level < zapcore.DebugLevel || level > zapcore.FatalLevel {
		level = zapcore.InfoLevel
	}

	// 编码器配置
	encoderCfg = zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var zapCore []zapcore.Core
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	// 控制台输出
	consoleWriter := zapcore.Lock(os.Stdout)
	zapCore = append(zapCore, zapcore.NewCore(
		encoder,
		consoleWriter,
		level,
	))
	// 文件输出配置
	fileWriter := zapcore.AddSync(logConfig)

	zapCore = append(zapCore, zapcore.NewCore(
		encoder,
		fileWriter,
		level,
	))

	// 编码器统一使用 JSON

	// 合并两个输出目标
	core := zapcore.NewTee(zapCore...)

	Log = zap.New(core, zap.AddCaller(),
		// zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

func Infof(format string, args ...interface{}) {
	if Log != nil {
		Log.Sugar().Infof(format, args...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

func Warnf(format string, args ...interface{}) {
	if Log != nil {
		Log.Sugar().Warnf(format, args...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

func Errorf(format string, args ...interface{}) {
	if Log != nil {
		Log.Sugar().Errorf(format, args...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

func Debugf(format string, args ...interface{}) {
	if Log != nil {
		Log.Sugar().Debugf(format, args)
	}
}

func GetLogger(m string) *zap.Logger {
	if Log == nil {
		// 默认配置
		loggerConf := map[string]interface{}{
			"filename":   "logs/app.log",
			"maxsize":    60,
			"maxbackups": 5,
			"maxage":     7,
			"compress":   true,
			"level":      -1,
		}
		level := loggerConf["level"]
		jsonBytes, err := json.Marshal(loggerConf)
		if err != nil {
			// 处理错误
			panic("Failed to marshal logger configuration: " + err.Error())
		}
		Init(jsonBytes, zapcore.Level(level.(int8))) // Initialize with default configuration if not already initialized
	}
	return Log.With(zap.String("module", m))
}
