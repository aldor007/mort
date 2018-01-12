package monitoring

import "go.uber.org/zap"

// logger is single singleton instance of logger
// default logger do nothing
var logger *zap.Logger = zap.NewNop()

// sugaredLogger extend version on zap.Logger that allow
// using sting format functions
var sugaredLogger *zap.SugaredLogger = logger.Sugar()

// RegisterLogger new logger as main logger for service
// RegisterLogger is NOT THREAD SAFE
func RegisterLogger(l *zap.Logger) {
	logger = l
}

// Log returns correct registered logger
func Log() *zap.Logger {
	return logger
}

// Logs return sugared zap logger
func Logs() *zap.SugaredLogger {
	return sugaredLogger
}
