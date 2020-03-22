package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

type capturedCrash error

func GetSentryOptions(name, version, commit string) sentry.ClientOptions {
	hostname, _ := os.Hostname()
	var sentryTransport sentry.Transport
	if os.Getenv("SENTRY_TRANSPORT") == "sync" {
		sentryTransport = sentry.NewHTTPSyncTransport()
	} else {
		sentryTransport = sentry.NewHTTPTransport()
	}

	return sentry.ClientOptions{
		AttachStacktrace: true,
		Transport:        sentryTransport,
		ServerName:       hostname,
		Release:          fmt.Sprintf("%s@%s [%s]", name, version, commit),
	}
}

func SentryDeferHandler() {
	defer sentry.Flush(2 * time.Second)

	r := recover()
	if r == nil {
		return
	}

	if _, ok := r.(capturedCrash); ok {
		fmt.Println("crash")
	} else {
		fmt.Println("unhandled panic")
		sentry.CaptureMessage(fmt.Sprintf("%v", r))
	}

	panic(r)
}

func CrashWithError(err error) {
	sentry.CaptureException(err)
	panic(capturedCrash(err))
}

func CrashWithMessage(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	sentry.CaptureException(err)
	panic(capturedCrash(err))
}

func CrashWithObject(obj interface{}) {
	if err, ok := obj.(error); ok {
		CrashWithError(err)
		return
	}

	CrashWithMessage("%v", obj)
}
