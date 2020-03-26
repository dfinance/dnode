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
	l := DNLogger{
		Logger: log.NewTMLogger(os.Stdout),
	}

	return &l
}

// Extended Error() implementation with Sentry support
func (l *DNLogger) Error(msg string, keyvals ...interface{}) {
	l.Logger.Error(msg, keyvals...)
	sentryCaptureMessage(msg, keyvals...)
}
