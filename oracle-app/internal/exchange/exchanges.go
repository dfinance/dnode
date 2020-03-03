package exchange

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	mu        sync.Mutex
	exchanges = make(map[string]Subscriber)

	o      sync.Once
	logger *logrus.Logger
)

func Register(name string, s Subscriber) {
	mu.Lock()
	exchanges[name] = s
	mu.Unlock()
}

func SetLogger(l *logrus.Logger) {
	o.Do(func() {
		logger = l
	})
}

func Logger() *logrus.Logger {
	return logger
}

func Exchanges() map[string]Subscriber {
	result := make(map[string]Subscriber)
	mu.Lock()
	for key, value := range exchanges {
		result[key] = value
	}
	mu.Unlock()

	return result
}
