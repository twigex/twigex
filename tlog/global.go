// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package tlog

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *Logger

func InitGlobalLogger(logger *Logger) {
	glob := *logger
	glob.zap = glob.zap.WithOptions(zap.AddCallerSkip(1))
	globalLogger = &glob
	Info = globalLogger.Info
	Warn = globalLogger.Warn
	Error = globalLogger.Error
}

type LogFunc func(format string, v ...interface{})

var Info LogFunc = defaultInfoLog
var Warn LogFunc = defaultWarnLog
var Error LogFunc = defaultErrorLog

func Infow(msg string, keysAndValues ...interface{}) {
	if globalLogger != nil {
		globalLogger.zap.Sugar().Infow(msg, keysAndValues...)
	} else {
		tempZapLogger().Sugar().Infow(msg, keysAndValues...)
	}
}

func Warnw(msg string, keysAndValues ...interface{}) {
	if globalLogger != nil {
		globalLogger.zap.Sugar().Warnw(msg, keysAndValues...)
	} else {
		tempZapLogger().Sugar().Warnw(msg, keysAndValues...)
	}
}

func Errorw(msg string, keysAndValues ...interface{}) {
	if globalLogger != nil {
		globalLogger.zap.Sugar().Errorw(msg, keysAndValues...)
	} else {
		tempZapLogger().Sugar().Errorw(msg, keysAndValues...)
	}
}

func defaultInfoLog(msg string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.zap.Sugar().Infof(msg, v...)
	} else {
		tempZapLogger().Sugar().Infof(msg, v...)
	}
}

func defaultWarnLog(msg string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.zap.Sugar().Warnf(msg, v...)
	} else {
		tempZapLogger().Sugar().Warnf(msg, v...)
	}
}

func defaultErrorLog(msg string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.zap.Sugar().Errorf(msg, v...)
	} else {
		tempZapLogger().Sugar().Errorf(msg, v...)
	}
}

func tempZapLogger() *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = zapcore.LowercaseLevelEncoder
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	return zap.New(core, zap.AddCaller())
}
