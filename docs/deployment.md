# Deployment

## Docker

```bash
docker run -p 8080:8080 arun0009/flakymock:latest
```

Mount a scenario file:

```bash
docker run -p 8080:8080 \
  -v $(pwd)/recipes/aws-s3-slow.yaml:/scenarios.yaml \
  arun0009/flakymock:latest
```

### Environment Variables

```bash
docker run -p 8080:8080 \
  -e RATE_LIMIT_RPS=10 \
  -e LOG_BODY=false \
  arun0009/flakymock:latest
```

See [Configuration](configuration.md) for all options.

## Docker Compose

```bash
docker compose up
```

## Local Development

Requires Go 1.26+.

```bash
git clone https://github.com/arun0009/flakymock.git
cd flakymock
go run main.go
```

## TLS (HTTPS)

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj '/CN=localhost'

export ENABLE_TLS=true
export CERT_FILE=./cert.pem
export KEY_FILE=./key.pem
go run main.go
```

## GitHub Actions

```yaml
- uses: arun0009/flakymock@v0.1.0
  with:
    port: 8080
    scenarios: ./tests/scenarios.yaml
```

The action installs FlakyMock, copies your scenario file, waits for `/health`, then leaves the server running for integration tests.
