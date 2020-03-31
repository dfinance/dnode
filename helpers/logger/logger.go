package logger

import (
	"os"

	"github.com/tendermint/tendermint/libs/log"
)

// Custom logger which extends standard Tendermint logger
type DNLogger struct {
	log.Logger
}

func NewDNLogger() log.Logger {
	return &DNLogger{
		Logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}
}

// Extended Error() implementation with Sentry support
func (l *DNLogger) Error(msg string, keyvals ...interface{}) {
	l.Logger.Error(msg, keyvals...)
	sentryCaptureMessage(msg, keyvals...)
}


// Method overwrite
func (l *DNLogger) With(keyvals ...interface{}) log.Logger {
	return &DNLogger{
		Logger: l.Logger.With(keyvals...),
	}
}
