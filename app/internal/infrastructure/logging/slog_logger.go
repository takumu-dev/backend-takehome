package logging

import (
	"context"
	"log/slog"
	"time"

	"blog-platform/internal/application/service"
)

// OperationLogger provides structured logging for application operations
type OperationLogger struct {
	logger *slog.Logger
}

// NewOperationLogger creates a new operation logger
func NewOperationLogger(logger *slog.Logger) *OperationLogger {
	return &OperationLogger{
		logger: logger,
	}
}

// Info logs an info message with context
func (l *OperationLogger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

// Error logs an error message with context
func (l *OperationLogger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// Warn logs a warning message with context
func (l *OperationLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

// Debug logs a debug message with context
func (l *OperationLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

// LogOperation logs the start and end of an operation with timing
func (l *OperationLogger) LogOperation(ctx context.Context, operation string, fn func() error) error {
	start := time.Now()
	l.Info(ctx, "operation started", "operation", operation)
	
	err := fn()
	duration := time.Since(start)
	
	if err != nil {
		l.Error(ctx, "operation failed", 
			"operation", operation, 
			"duration", duration,
			"error", err.Error())
	} else {
		l.Info(ctx, "operation completed", 
			"operation", operation, 
			"duration", duration)
	}
	
	return err
}

// LogOperationWithResult logs an operation that returns a result
func (l *OperationLogger) LogOperationWithResult(ctx context.Context, operation string, fn func() (interface{}, error)) (interface{}, error) {
	start := time.Now()
	l.Info(ctx, "operation started", "operation", operation)
	
	result, err := fn()
	duration := time.Since(start)
	
	if err != nil {
		l.Error(ctx, "operation failed", 
			"operation", operation, 
			"duration", duration,
			"error", err.Error())
	} else {
		l.Info(ctx, "operation completed", 
			"operation", operation, 
			"duration", duration)
	}
	
	return result, err
}

// Verify that OperationLogger implements the Logger interface
var _ service.Logger = (*OperationLogger)(nil)
