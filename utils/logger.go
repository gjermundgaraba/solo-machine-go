package utils

import (
	"cosmossdk.io/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type SDKLoggerWrapper struct {
	zapLogger *zap.Logger
}

func NewSDKLoggerWrapper(logger *zap.Logger) SDKLoggerWrapper {
	return SDKLoggerWrapper{
		zapLogger: logger,
	}
}

func (s SDKLoggerWrapper) Info(msg string, keyVals ...any) {
	s.zapLogger.Info(msg, zap.Any("vals", keyVals))
}

func (s SDKLoggerWrapper) Warn(msg string, keyVals ...any) {
	s.zapLogger.Warn(msg, zap.Any("vals", keyVals))
}

func (s SDKLoggerWrapper) Error(msg string, keyVals ...any) {
	s.zapLogger.Error(msg, zap.Any("vals", keyVals))
}

func (s SDKLoggerWrapper) Debug(msg string, keyVals ...any) {
	s.zapLogger.Debug(msg, zap.Any("vals", keyVals))
}

func (s SDKLoggerWrapper) With(keyVals ...any) log.Logger {
	return s
}

func (s SDKLoggerWrapper) Impl() any {
	return s.zapLogger
}

func CreateLogger(verbose bool) *zap.Logger {
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerConfig.Encoding = "console"
	loggerConfig.Level = zap.NewAtomicLevelAt(logLevel)

	// Create the logger from the core
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
