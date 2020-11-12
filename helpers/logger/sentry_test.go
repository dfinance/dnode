// +build integ_sentry

package logger

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	sentryApi "github.com/atlassian/go-sentry-api"
	"github.com/stretchr/testify/require"
)

func Test_SentryIntegration(t *testing.T) {
	const (
		defSentryUrl          = "https://sentry.dfinance.co/api/0/"
		defSentryOrganisation = "sentry"
		defSentryProject      = "integ-testing"
		defSentryEnvironment  = "testing"
		defSentryConTimeoutS  = 5
		defSentryPollTimeoutS = 10
		errMsgFmt             = "dnode Sentry integ test msg #%d"
	)

	// check inputs
	require.NoError(t, os.Setenv("DN_SENTRY_ENVIRONMENT", defSentryEnvironment), "set env")
	if os.Getenv("DN_SENTRY_DSN") == "" {
		t.Fatalf("%q env: empty", "DN_SENTRY_DSN")
	}
	inputSentryToken := os.Getenv("DN_SENTRY_TEST_TOKEN")
	if inputSentryToken == "" {
		t.Fatalf("%q env: empty", "DN_SENTRY_TEST_TOKEN")
	}
	inputSentryUrl := defSentryUrl
	if env := os.Getenv("DN_SENTRY_TEST_URL"); env != "" {
		inputSentryUrl = env
	}
	inputSentryOrg := defSentryOrganisation
	if env := os.Getenv("DN_SENTRY_TEST_ORG"); env != "" {
		inputSentryOrg = env
	}
	inputSentryPrj := defSentryProject
	if env := os.Getenv("DN_SENTRY_TEST_PRJ"); env != "" {
		inputSentryPrj = env
	}

	// setup Sentry
	require.NoError(t, SetupSentry("dnode", "vx.x.x", "_"), "Sentry init")

	// setup Sentry client
	sentryConTimeout := defSentryConTimeoutS
	sentryClient, err := sentryApi.NewClient(inputSentryToken, &inputSentryUrl, &sentryConTimeout)
	require.NoError(t, err, "create Sentry client")

	sentryOrg, err := sentryClient.GetOrganization(inputSentryOrg)
	require.NoError(t, err, "get Sentry organization")

	sentryPrj, err := sentryClient.GetProject(sentryOrg, inputSentryPrj)
	require.NoError(t, err, "gGet Sentry project")

	// prepare and send error message
	logger := NewDNLogger()
	rand.Seed(time.Now().UnixNano())
	errMsg := fmt.Sprintf(errMsgFmt, rand.Int())

	errMsgSendAt := time.Now()
	logger.Error(errMsg)

	// poll project issues searching for errMsg
	timeoutDur := defSentryPollTimeoutS * time.Second
	timeoutCh := time.After(timeoutDur)
	targetIssue := sentryApi.Issue{}
	for {
		time.Sleep(100 * time.Millisecond)

		issues, _, err := sentryClient.GetIssues(sentryOrg, sentryPrj, nil, nil, nil)
		require.NoError(t, err, "get issues")

		found := false
		for _, issue := range issues {
			require.NotNil(t, issue.Title, "issue.Title")
			if *issue.Title != errMsg {
				continue
			}

			events, _, err := sentryClient.GetIssueEvents(issue)
			require.NoError(t, err, "get issue events")
			require.Len(t, events, 1, "issue events length")
			event := events[0]

			require.NotNil(t, event.DateCreated, "event.DateCreated")
			if (*event.DateCreated).After(errMsgSendAt) {
				continue
			}

			envTagChecked := false
			require.NotNil(t, event.Tags, "event.Tags")
			for i, tag := range *event.Tags {
				require.NotNil(t, tag.Key, "event.Tag.Key %d", i)
				require.NotNil(t, tag.Value, "event.Tag.Value %d", i)
				if *(tag.Key) == "environment" && *(tag.Value) == defSentryEnvironment {
					envTagChecked = true
				}
			}
			require.True(t, envTagChecked, "tag %q not found in event.Tags", "environment")

			targetIssue, found = issue, true
			break
		}

		if found {
			break
		}

		select {
		case <-timeoutCh:
			t.Fatalf("issue with msg %q not found after %v", errMsg, timeoutDur)
		default:
		}
	}

	// remove issue
	require.NoError(t, sentryClient.DeleteIssue(targetIssue), "remove issue")
}
