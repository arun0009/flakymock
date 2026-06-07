# FlakyMock vs WireMock vs Mountebank

This is an honest comparison. **FlakyMock is not trying to replace every mock server.** It is built for a specific job: **mock APIs and test resilience behavior in one tool.**

---

## One-line summary

| Tool | Best for |
|------|----------|
| **WireMock** | General-purpose HTTP API mocking at org scale |
| **Mountebank** | Multi-protocol mocking (HTTP, TCP, SMTP, etc.) |
| **FlakyMock** | HTTP mocking **plus** chaos/resilience testing (retries, timeouts, circuit breakers) |

---

## What each tool optimizes for

### WireMock
- Stub any request → return any response
- Record and playback from real traffic
- Rich admin API, persistence, verification
- Huge ecosystem, language bindings, WireMock Cloud

**Mental model:** "I need a fake API that behaves predictably."

### Mountebank
- Mock multiple protocols, not just HTTP
- Imposters per port, proxy mode, programmatic test DSL

**Mental model:** "I need a programmable mock platform for many service types."

### FlakyMock
- Stateful fault sequences ("fail twice, then succeed")
- Built-in chaos: delay jitter, probability faults, circuit breaker simulation
- Header-driven faults without config (`X-Echo-Delay`, `X-Echo-Status`)
- Prometheus fault metrics, CPU/memory stress endpoints

**Mental model:** "I need a flaky downstream so I can test retries, timeouts, and circuit breakers."

---

## Feature comparison

| Capability | WireMock | Mountebank | FlakyMock |
|------------|----------|------------|-------------------|
| HTTP stubbing | ✅ | ✅ | ✅ |
| TCP / SMTP / other protocols | ❌ | ✅ | ❌ |
| Record & playback | ✅ | ✅ (varies) | ❌ (history + replay only) |
| Persistence across restarts | ✅ (files) | ✅ | ✅ (`SCENARIO_ROOT_DIR/mappings/`) |
| Request verification | ✅ | ✅ | ✅ (count, body, headers, interval, sequence) |
| Stateful response sequences | ⚠️ (via scenarios/state) | ⚠️ | ✅ (first-class) |
| Probability-based faults | ⚠️ (possible, not core) | ⚠️ | ✅ (first-class) |
| Delay jitter (`100ms-500ms`) | ⚠️ | ⚠️ | ✅ |
| Circuit breaker simulation | ❌ | ❌ | ✅ |
| Header-driven chaos (no YAML) | ❌ | ❌ | ✅ |
| CPU / memory stress endpoints | ❌ | ❌ | ✅ |
| Prometheus fault metrics | ⚠️ (custom) | ❌ | ✅ |
| Single binary + Docker | ✅ | ✅ (Node) | ✅ (Go) |

✅ = strong / built-in · ⚠️ = possible with extra work · ❌ = not a focus today

---

## When to choose FlakyMock

Choose it when you are testing **client-side resilience**, not just response shapes:

- "Does my app **retry** after two 503s?"
- "Does my client **respect Retry-After** on 429?"
- "Does my **circuit breaker open** after repeated failures?"
- "Does my timeout logic survive **200–500ms jitter**?"
- "What happens when the downstream is **slow and flaky**?"

You can do some of this with WireMock or Mountebank, but it is not their primary story. You often end up bolting chaos on top of a plain mock.

---

## When to stick with WireMock or Mountebank

Stick with **WireMock** if you need:
- Org-wide standard mock server
- Record/playback from production-like traffic
- Mature admin API, persistence, and verification today
- Non-Go polyglot teams with existing WireMock investment

Stick with **Mountebank** if you need:
- Non-HTTP protocols
- One tool mocking SMTP, TCP, and HTTP together

---

## Side-by-side: retry test

**Goal:** First call returns 503, second call returns 200. Assert the client retried.

### WireMock (conceptual)

```json
{
  "scenarioName": "retry-test",
  "requiredScenarioState": "Started",
  "newScenarioState": "Failed-once",
  "request": { "method": "GET", "url": "/users" },
  "response": { "status": 503 }
}
```

Plus a second stub for state `Failed-once` → 200, and WireMock's verify API to assert call count.

### FlakyMock

```yaml
- path: /users
  method: GET
  responses:
    - status: 503
      body: '{"error": "unavailable"}'
    - status: 200
      body: '{"status": "ok"}'
```

Or inject at runtime:

```bash
curl -X POST http://localhost:8080/scenario \
  -H "Content-Type: application/json" \
  -d '[{"path":"/users","method":"GET","responses":[{"status":503},{"status":200}]}]'
```

**Difference:** fault sequences and chaos knobs are the default workflow, not an advanced WireMock scenario-state pattern.

---

## Side-by-side: zero-config chaos

**Goal:** Add 500ms delay and return 503 — no YAML, no restart.

### WireMock / Mountebank
Requires a stub mapping (file or admin API). No built-in "chaos headers on echo" pattern.

### FlakyMock

```bash
curl -i http://localhost:8080/echo \
  -H "X-Echo-Delay: 500ms" \
  -H "X-Echo-Status: 503"
```

Useful for quick experiments, manual testing, and debugging retry logic without writing config.

---

## What FlakyMock still needs (honest gaps)

1. **Record/playback** — create stubs from live traffic
2. **Proxy pass-through** — forward to real backend when no stub matches
3. **Namespaces** — per-team scenario directories

**Recently shipped:** file persistence, scenario CRUD, richer verification (body, headers, retry interval, sequence).

See [Traction Plan](traction-plan.md) for the prioritized roadmap.

---

## Bottom line

- **WireMock** mocks.
- **Mountebank** mocks across protocols.
- **FlakyMock** mocks **and breaks things on purpose** so you can test resilience in one place.

If your problem is "stub this API response," use WireMock. If your problem is "prove my service survives a bad downstream," FlakyMock is the better fit — once verification and CI ergonomics catch up.
