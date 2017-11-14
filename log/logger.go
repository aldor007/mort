package log

import "go.uber.org/zap"


// logger is single singleton instance of logger
// defaalt logger do nothing
var logger *zap.SugaredLogger = zap.NewNop().Sugar()

// RetiserLogger register new logger as main logger for service
// RegisterLogger is NOT THREAD SAFE
func RegisterLogger(l *zap.SugaredLogger) {
	logger = l
}

// Log returns correct registered logger
func Log() *zap.SugaredLogger {
	return logger
}