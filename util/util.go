package util

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CreateLogger creates either a dev or prod zap logger
//
// Dev logger: level = DEBUG, colored output, no sampling, stack traces incl. for WARN+ messages
// Prod logger: level = INFO, JSON output to stderr, sampling, no stack traces
func CreateLogger(prod bool, fields map[string]interface{}) (*zap.Logger, error) {
	var loggerConfig zap.Config

	if prod {
		loggerConfig = zap.NewProductionConfig()
	} else {
		loggerConfig = zap.NewDevelopmentConfig()
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	loggerConfig.InitialFields = fields

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
