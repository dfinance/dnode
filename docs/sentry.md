# Sentry integration

[Sentry](https://sentry.io) is used to capture crash reports for `dnode` and `dncli` applications.
By default the integration is switched off.

The following environment variables should be defined in order to enable the integration:
* `DN_SENTRY_DSN` - Sentry DSN token (`https://[token]@sentry.io/5167345`);
* `DN_SENTRY_ENVIRONMENT` - sets the environment code to separate events from testnet and production (could be empty);
