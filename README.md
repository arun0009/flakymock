<div align="center">
  <img src="docs/flakymock.png" alt="FlakyMock Mascot" width="200"/>
</div>

<h1 align="center">FlakyMock</h1>

<p align="center">
  <strong>Stub the API. Break it on purpose. Assert your client handled it.</strong><br>
  One HTTP mock server for integration tests — delays, 503s, 429s, circuit breakers, and call verification.
</p>

<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/arun0009/flakymock)](https://goreportcard.com/report/github.com/arun0009/flakymock)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/arun0009/flakymock)](https://golang.org/)

</div>

---

## The problem

WireMock is great at returning the response you asked for. **FlakyMock is for when the downstream misbehaves.**

Run it as a Docker sidecar and pretend to be Stripe, S3, or a legacy API that flakes — then prove your app retried, backed off, or tripped a circuit breaker. No second chaos tool. Any language can hit it over HTTP.

> Not sure if this fits your stack? See [FlakyMock vs WireMock](docs/comparison.md).

---

## Quick start

```bash
docker run -p 8080:8080 arun0009/flakymock:latest
```

**Zero-config fault** — no YAML, just headers:

```bash
curl -i "http://localhost:8080/echo" \
  -H "X-Echo-Delay: 500ms" \
  -H "X-Echo-Status: 503"
```

**Stateful retry sequence** — fail twice, then succeed:

```bash
curl -X POST http://localhost:8080/scenario \
  -H "Content-Type: application/json" \
  -d '[{"path":"/users","method":"GET","responses":[{"status":503},{"status":503},{"status":200}]}]'

curl -i http://localhost:8080/users   # 503
curl -i http://localhost:8080/users   # 503
curl -i http://localhost:8080/users   # 200
```

**Assert retries in tests** — did the client call enough times?

```bash
curl -s "http://localhost:8080/api/verify/requests?path=/users&method=GET&min=3"
# {"count":3,"matched":true,...}
```

**Reset for the next test run:**

```bash
curl -X POST "http://localhost:8080/api/control/reset-scenarios?path=/users&method=GET"
```

Full walkthrough: **[Resilience Testing Example](docs/resilience-testing-example.md)** (~10 min).

---

## When to use FlakyMock

| Choose **FlakyMock** | Choose **WireMock** |
|----------------------|---------------------|
| Retries, timeouts, circuit breakers | Org-wide general API mocking |
| Stateful sequences (fail N, then succeed) | Record/playback from real traffic |
| Header-driven faults without config | Mature ecosystem and cloud offering |
| Built-in verify + reset for CI | You already run WireMock everywhere |

---

## Test in CI

Copy-paste integration test with [testcontainers-go](examples/testcontainers/):

```bash
cd examples/testcontainers
go test -tags=integration -v .
```

Starts FlakyMock from the repo Dockerfile, runs a 503→503→200 scenario, verifies call count, resets state.

---

## Recipes

Mount a file and go — pre-built scenarios in `recipes/`:

```bash
# Slow / flaky storage
docker run -p 8080:8080 \
  -v $(pwd)/recipes/aws-s3-slow.yaml:/scenarios.yaml \
  arun0009/flakymock:latest

# Circuit breaker: healthy → spike → recovery
# recipes/generic-circuit-breaker.yaml

# Rate limit with Retry-After
# recipes/stripe-rate-limit.yaml
```

---

## Features

**Mocking**
- Path, method, header, query, and JSON body matching
- Dynamic scenarios via `POST /scenario`; YAML on startup
- File persistence under `SCENARIO_ROOT_DIR/mappings/`

**Fault injection**
- Stateful response sequences and probability-based faults
- Delay jitter (`delayRange: "200ms-500ms"`)
- Circuit breaker simulation (closed → open → half-open)
- Chaos headers on `/echo` (`X-Echo-Delay`, `X-Echo-Status`)

**Test assertions**
- `GET /api/verify/requests` — count calls, filter by body/headers/status
- `POST /api/verify/sequence` — assert call order
- `POST /api/control/reset-scenarios` — repeatable CI runs

**Ops**
- Prometheus metrics at `/metrics`
- Request history, replay, WebSocket/SSE testers
- CPU/memory stress endpoints for neighbor testing

---

## Install

```bash
# Docker (recommended)
docker run -p 8080:8080 arun0009/flakymock:latest

# Go
go install github.com/arun0009/flakymock@latest
```

---

## Docs

| Doc | What you'll find |
|-----|------------------|
| [Resilience Testing Example](docs/resilience-testing-example.md) | End-to-end retry demo |
| [Comparison](docs/comparison.md) | FlakyMock vs WireMock |
| [Scenarios](docs/scenarios.md) | YAML reference |
| [API Reference](docs/api_reference.md) | Endpoints |
| [Configuration](docs/configuration.md) | Environment variables |
