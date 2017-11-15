package log

import "go.uber.org/zap"


// logger is single singleton instance of logger
// default logger do nothing
var logger *zap.Logger = zap.NewNop()
var suggaredLogger *zap.SugaredLogger = logger.Sugar()

// RetiserLogger register new logger as main logger for service
// RegisterLogger is NOT THREAD SAFE
func RegisterLogger(l *zap.Logger) {
	logger = l
}

// Log returns correct registered logger
func Log() *zap.Logger {
	return logger
}

// Logs return suggared zap logger
func Logs() *zap.SugaredLogger  {
	return suggaredLogger
}