package solomachine

import (
	"cosmossdk.io/log"
	"go.uber.org/zap"
)

type sdkloggerwrapper struct {
	zapLogger *zap.Logger
}

func (s sdkloggerwrapper) Info(msg string, keyVals ...any) {
	s.zapLogger.Info(msg, zap.Any("vals", keyVals))
}

func (s sdkloggerwrapper) Warn(msg string, keyVals ...any) {
	s.zapLogger.Warn(msg, zap.Any("vals", keyVals))
}

func (s sdkloggerwrapper) Error(msg string, keyVals ...any) {
	s.zapLogger.Error(msg, zap.Any("vals", keyVals))
}

func (s sdkloggerwrapper) Debug(msg string, keyVals ...any) {
	s.zapLogger.Debug(msg, zap.Any("vals", keyVals))
}

func (s sdkloggerwrapper) With(keyVals ...any) log.Logger {
	return s
}

func (s sdkloggerwrapper) Impl() any {
	return s.zapLogger
}
