package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/icoder-new/avito-shop/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.Logger
}

func New(cfg config.LoggerSettings) (*Logger, error) {
	if err := os.MkdirAll(cfg.LogFile, 0744); err != nil {
		return nil, fmt.Errorf("cannot create log directory: %w", err)
	}

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	logLevel := getLogLevel(cfg.Level)
	logPath := filepath.Join(cfg.LogFile, fmt.Sprintf("avito-shop-%s.log", time.Now().Format("2006-01-02")))

	jsonConfig := zap.NewProductionEncoderConfig()
	jsonConfig.TimeKey = "timestamp"
	jsonConfig.EncodeTime = zapcore.TimeEncoderOfLayout(cfg.TimeFormat)
	jsonEncoder := zapcore.NewJSONEncoder(jsonConfig)

	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleConfig.EncodeTime = zapcore.TimeEncoderOfLayout(cfg.TimeFormat)
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, fileWriter, logLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel),
	)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(
			zap.Int("pid", os.Getpid()),
			zap.Int("go_routines", runtime.NumGoroutine()),
			zap.String("log_file", logPath),
		),
	)

	return &Logger{logger}, nil
}

func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
