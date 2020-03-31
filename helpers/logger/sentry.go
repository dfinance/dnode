package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

var sentryCfg sentryConfig

type sentryConfig struct {
	enabled         bool
	logSends        bool
	dsnToken        string
	environmentCode string
	hostname        string
	appName         string
	appVersion      string
	appCommit       string
	sendTimeout     time.Duration
}

func (c sentryConfig) getClientOptions() sentry.ClientOptions {
	sentryTransport := sentry.NewHTTPSyncTransport()
	sentryTransport.Timeout = c.sendTimeout

	return sentry.ClientOptions{
		AttachStacktrace: true,
		Transport:        sentryTransport,
		Dsn:              c.dsnToken,
		Environment:      c.environmentCode,
		ServerName:       c.hostname,
		Release:          fmt.Sprintf("%s@%s [%s]", c.appName, c.appVersion, c.appCommit),
	}
}

func SetupSentry(appName, appVersion, appCommit string) error {
	// force overwrite standard Sentry envs
	if err := os.Setenv("SENTRY_DSN", ""); err != nil {
		return fmt.Errorf("can't overwrite %q: %w", "SENTRY_DSN", err)
	}
	if err := os.Setenv("SENTRY_ENVIRONMENT", ""); err != nil {
		return fmt.Errorf("can't overwrite %q: %w", "SENTRY_ENVIRONMENT", err)
	}

	sentryDsn := os.Getenv("DN_SENTRY_DSN")
	sentryEnvironment := os.Getenv("DN_SENTRY_ENVIRONMENT")
	hostname, _ := os.Hostname()

	if appName == "" {
		appName = "undefined"
	}
	if appVersion == "" {
		appVersion = "v0.0.0"
	}
	if sentryEnvironment == "" {
		sentryEnvironment = "undefined"
	}

	sentryCfg.enabled = false
	sentryCfg.logSends = true
	sentryCfg.dsnToken = sentryDsn
	sentryCfg.environmentCode = sentryEnvironment
	sentryCfg.appName = appName
	sentryCfg.appVersion = appVersion
	sentryCfg.appCommit = appCommit
	sentryCfg.hostname = hostname
	sentryCfg.sendTimeout = 2 * time.Second

	if sentryCfg.dsnToken != "" {
		if err := sentry.Init(sentryCfg.getClientOptions()); err != nil {
			return fmt.Errorf("sentry init: %w", err)
		}
		sentryCfg.enabled = true
	}

	return nil
}

func CrashDeferHandler() {
	r := recover()
	if r == nil {
		return
	}

	sentryCaptureObject(r)
	panic(r)
}

func sentryCaptureMessage(format string, args ...interface{}) {
	if !sentryCfg.enabled {
		sentryLogSend(nil)
		return
	}

	msg := fmt.Sprintf(format, args...)
	sentryLogSend(sentry.CaptureMessage(msg))
}

func sentryCaptureObject(obj interface{}) {
	if !sentryCfg.enabled {
		sentryLogSend(nil)
		return
	}

	err := fmt.Errorf("%T: %v", obj, obj)
	sentryLogSend(sentry.CaptureException(err))
}

func sentryLogSend(eventId *sentry.EventID) {
	if !sentryCfg.logSends {
		return
	}

	if !sentryCfg.enabled {
		fmt.Println("sentry send: skipped")
		return
	}

	if eventId == nil {
		fmt.Println("sentry send: failed")
		return
	}

	fmt.Println("sentry send: done")
}
