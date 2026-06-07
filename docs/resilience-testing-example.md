# Resilience Testing Example: Retry on 503

This is the **killer demo** for FlakyMock: prove your HTTP client retries when a downstream fails, then recovers.

**Time:** ~10 minutes  
**Prerequisites:** Docker (or `go run main.go`)

---

## What we are testing

A downstream API that:
1. Returns **503** on the first call
2. Returns **503** on the second call
3. Returns **200** on the third call

Your client should retry and eventually succeed. FlakyMock models that flaky downstream in one YAML (or one `POST /scenario`).

---

## Step 1: Start the mock

```bash
docker run -p 8080:8080 arun0009/flakymock:latest
```

---

## Step 2: Inject the retry scenario

```bash
curl -X POST http://localhost:8080/scenario \
  -H "Content-Type: application/json" \
  -d '[
    {
      "path": "/payments",
      "method": "POST",
      "responses": [
        {"status": 503, "body": "{\"error\": \"unavailable\"}"},
        {"status": 503, "body": "{\"error\": \"unavailable\"}"},
        {"status": 200, "body": "{\"status\": \"paid\"}"}
      ]
    }
  ]'
```

---

## Step 3: Simulate a client with retries

Save as `retry_client.go` (or run mentally with curl three times):

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{Timeout: 2 * time.Second}
	url := "http://localhost:8080/payments"

	var lastStatus int
	for attempt := 1; attempt <= 5; attempt++ {
		resp, err := client.Post(url, "application/json", nil)
		if err != nil {
			fmt.Printf("attempt %d: error: %v\n", attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		lastStatus = resp.StatusCode
		fmt.Printf("attempt %d: status=%d body=%s\n", attempt, resp.StatusCode, body)
		if resp.StatusCode == 200 {
			fmt.Println("success after retries")
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("gave up; last status=%d\n", lastStatus)
}
```

Run:

```bash
go run retry_client.go
```

**Expected output:**

```
attempt 1: status=503 body={"error": "unavailable"}
attempt 2: status=503 body={"error": "unavailable"}
attempt 3: status=200 body={"status": "paid"}
success after retries
```

---

## Step 4: Verify the mock saw 3 calls

```bash
curl -s "http://localhost:8080/api/verify/requests?path=/payments&method=POST&min=3"
# {"path":"/payments","method":"POST","count":3,"min":3,"matched":true}
```

---

## Step 5: Reset for the next test run

Stateful scenarios advance on each request. Reset before each test so CI stays deterministic:

```bash
curl -X POST "http://localhost:8080/api/control/reset-scenarios?path=/payments&method=POST"
# {"reset":1,"path":"/payments","method":"POST"}
```

---

## Same test with a recipe file

Save as `retry-scenario.yaml`:

```yaml
- path: /payments
  method: POST
  responses:
    - status: 503
      body: '{"error": "unavailable"}'
    - status: 503
      body: '{"error": "unavailable"}'
    - status: 200
      body: '{"status": "paid"}'
```

Run with the recipe mounted:

```bash
docker run -p 8080:8080 \
  -v $(pwd)/retry-scenario.yaml:/scenarios.yaml \
  arun0009/flakymock:latest
```

---

## Circuit breaker variant

Use the built-in recipe:

```bash
docker run -p 8080:8080 \
  -v $(pwd)/recipes/generic-circuit-breaker.yaml:/scenarios.yaml \
  arun0009/flakymock:latest
```

Then hit `GET /backend/resource` repeatedly and watch the sequence: healthy → failures → recovery. See [Scenarios](scenarios.md#circuit-breaker) for details.

---

## Next steps

- [Comparison with WireMock](comparison.md)
- [Scenarios](scenarios.md) — full YAML reference
- [API Reference](api_reference.md) — verify and reset endpoints
