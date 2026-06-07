# FlakyMock + Testcontainers

End-to-end integration test: start FlakyMock in Docker, inject a retry scenario, verify call count, reset state.

## Prerequisites

- Docker running locally

## Run

```bash
cd examples/testcontainers
go mod tidy
go test -tags=integration -v .
```

The test builds FlakyMock from the repo `Dockerfile`, posts a 503→503→200 scenario, asserts the client recovers, verifies request count via `/api/verify/requests`, and resets via `/api/control/reset-scenarios`.
