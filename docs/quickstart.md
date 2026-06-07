# Quickstart

**Docker (recommended):**

```bash
docker run -p 8080:8080 arun0009/flakymock:latest
```

**Inject a fault with no config:**

```bash
curl -i http://localhost:8080/echo \
  -H "X-Echo-Delay: 500ms" \
  -H "X-Echo-Status: 503"
```

**Full retry walkthrough:** [Resilience Testing Example](resilience-testing-example.md) (~10 min).

**From source (Go 1.26+):**

```bash
git clone https://github.com/arun0009/flakymock.git
cd flakymock
go run main.go
```
