# fetch-jwks

A Go command-line tool for fetching and caching JWKS documents from OAuth2/OIDC issuers. It supports retries with backoff and conditional requests via ETag/If-None-Match to avoid unnecessary downloads.

## Getting Started

- Build: `make build`
- Test: `make test`
- Format: `make fmt`
- Lint: `make lint`

Requires Go 1.22+.

## Pre-commit

Install hooks once per clone: `pre-commit install`. Run them manually with `pre-commit run --all-files`. Hooks include gofmt, golangci-lint (--fast), and basic whitespace checks.

## Configuration

See [examples/fetch-jwks.example.yaml](examples/fetch-jwks.example.yaml) for a minimal config. You can also specify issuers via repeatable `-issuer issuer=<url>,jwks_uri=<url>` flags.
