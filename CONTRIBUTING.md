# Contributing to go-squad

Thank you for helping improve this SDK. This document covers how to contribute effectively.

## Getting Started

```bash
git clone https://github.com/kingztech2019/go-squad
cd go-squad
go mod download
make test
```

## Development Requirements

- Go 1.21 or later
- `golangci-lint` for linting (`brew install golangci-lint` or see [installation](https://golangci-lint.run/welcome/install/))

## Making Changes

### Adding a new API method

1. Add the request/response structs to the relevant `*_types.go` file.
2. Add the method to the service in the corresponding `*.go` file.
3. Add tests in `*_test.go` using `newTestServer` and `newTestClient` from `testhelpers_test.go`.
4. Ensure the method accepts `context.Context` as the first argument.
5. Amounts must use `int64` (kobo), never `float64`.

### Adding a new service

1. Create `myservice.go` and `myservice_types.go` in the package root.
2. Add a field to `Client` in [squad.go](squad.go).
3. Wire the service in `New()` in [squad.go](squad.go).
4. Add tests in `myservice_test.go`.

## Testing

```bash
# Run all tests with the race detector
make test

# Quick run without race detector
go test ./...

# Single test
go test -run TestInitiatePayment_Success ./...
```

All tests use `net/http/httptest` to mock the Squad API — no real network calls, no sandbox key required.

## Code Style

```bash
make fmt   # format code
make vet   # run go vet
make lint  # run golangci-lint (requires installation)
```

Key conventions:
- All public API methods accept `context.Context` as the first argument
- All monetary amounts are `int64` in the lowest denomination (kobo for NGN)
- No external runtime dependencies — use the standard library only
- File upload APIs accept `[]byte`, not `*os.File`

## Pull Request Process

1. Fork the repo and create a feature branch from `main`.
2. Make your changes with tests.
3. Run `make test` and `make vet` — both must pass.
4. Update `CHANGELOG.md` under `[Unreleased]`.
5. Open a pull request with a clear description of the change.

## Reporting Issues

Open a GitHub issue with:
- The SDK version
- A minimal code snippet reproducing the problem
- The Squad API response (redact your secret key)

## Licence

By contributing you agree that your contributions will be licensed under the [MIT License](LICENSE).
