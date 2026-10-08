// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package tlog

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.Logger
}

// Set up the encoder configuration (JSON output)
func newConfig() *zapcore.EncoderConfig {
	config := zap.NewProductionEncoderConfig()

	// Ensure that log levels are lowercase in logs
	config.EncodeLevel = zapcore.LowercaseLevelEncoder
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	return &config
}

func NewLogger() *Logger {
	l := &Logger{}

	config := newConfig()
	if err := os.MkdirAll("logs", 0700); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create logs directory:", err)
	}

	// JSON Encoder for log file and console (everything in JSON)
	jsonEncoder := zapcore.NewJSONEncoder(*config)

	// Open log file for writing
	logFile, _ := os.OpenFile("logs/twigex.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)

	// Set the log level
	defaultLogLevel := zapcore.DebugLevel

	// Set up the core to log to both file (JSON) and console (JSON)
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, writer, defaultLogLevel),                     // JSON format for file
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel), // JSON format for console
	)

	// Create a logger instance
	logger := zap.New(core, zap.AddCaller())

	l.zap = logger

	return l
}

// Error level logging
func (l *Logger) Error(format string, v ...interface{}) {
	l.zap.Error(fmt.Sprintf(format, v...))
}

// Info level logging
func (l *Logger) Info(format string, v ...interface{}) {
	l.zap.Info(fmt.Sprintf(format, v...))
}

// Warn level logging
func (l *Logger) Warn(format string, v ...interface{}) {
	l.zap.Warn(fmt.Sprintf(format, v...))
}
