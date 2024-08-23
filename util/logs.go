package util

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// LoggerFromContext returns a logr.Logger from the context or an instance of the global logger
func LoggerFromContext(ctx context.Context) logr.Logger {
	l, err := logr.FromContext(ctx)
	if err != nil {
		// We need to AllowZapFields to ensure the protobuf message is logged correctly as a json object.
		// For that to work we need to do logr.Info("message", zap.Object("key", protoMessage))
		// Which means we are passing zap.Field to the logr interface.
		return zapr.NewLoggerWithOptions(zap.L(), zapr.AllowZapFields(true))
	}
	return l
}
