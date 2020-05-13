# Sentry integration

[Sentry](https://sentry.io) is used to capture crash reports for `dnode` and `dncli` applications.
By default the integration is switched off.

The following environment variables should be defined in order to enable the integration:
* `DN_SENTRY_DSN` - Sentry DSN token (`https://[token]@sentry.io/5167345`);
* `DN_SENTRY_ENVIRONMENT` - sets the environment code to separate events from testnet and production (could be empty);

## Testing

Sentry integration test can be executed with the following command:

    go test ./... --tags=integ_sentry

The following environment variables should be defined:
* `DN_SENTRY_DSN` - DSN token;
* `DN_SENTRY_TEST_URL` - API URL (`https://sentry.{domain}/api/0/`);
* `DN_SENTRY_TEST_TOKEN` - test-user auth token;
* `DN_SENTRY_TEST_ORG` - organization code;
* `DN_SENTRY_TEST_PRJ` - project code;
