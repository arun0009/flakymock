# Traction Plan

A concrete plan to grow adoption for FlakyMock. The goal is not "more features for their own sake" — it is to make the **mock + resilience** story obvious, provable, and easy to adopt.

---

## Positioning (use everywhere)

**Tagline:** Mock APIs and test resilience in one tool.

**Pitch:** WireMock stubs responses. FlakyMock stubs responses *and* injects faults — delays, 503s, 429s, circuit breaker behavior — so you can test retries, timeouts, and fallbacks without a separate chaos tool.

**Target user:** Backend / platform engineers writing integration tests for HTTP clients, especially Go teams using Docker or testcontainers in CI.

**Do not claim:** "Better than WireMock at everything." Claim: "Better for resilience and chaos testing."

---

## Phase 1 — Prove the value (weeks 1–2)

### 1.1 Ship one killer example
**Doc:** [Resilience Testing Example](resilience-testing-example.md)

One end-to-end walkthrough:
1. Start the mock (Docker)
2. Load a retry scenario (503 → 503 → 200)
3. Point a small HTTP client at it
4. Show the client retried and succeeded
5. (Interim) Use `/history` to count calls until verification API exists

**Done when:** A new user can copy-paste and see resilience testing work in under 10 minutes.

### 1.2 Update README and GitHub metadata
- [x] Sharper README opening (mock + resilience, not generic chaos)
- [ ] GitHub repo description: `Mock server + resilience testing — retries, circuit breakers, latency, rate limits`
- [ ] Topics: keep existing; add `integration-testing`, `retry`, `circuit-breaker`, `testcontainers`
- [ ] Pin the resilience example in README

### 1.3 Publish once
Pick one channel and post the killer example:
- Show HN (`Show HN: FlakyMock – mock server built for retry/circuit breaker testing`)
- r/golang or r/devops
- Dev.to / personal blog with the same content

**Done when:** One public URL points at the repo with a working demo.

---

## Phase 2 — Remove adoption blockers (weeks 3–6)

Implement these three features first. They matter more than persistence for early traction.

### Feature 1: Request verification API ✅

**Shipped:** `GET /api/verify/requests`

**Why:** Injecting faults is half the story. Tests need to assert "my client retried 3 times."

**Proposed API:**

```
GET /api/verify/requests?path=/users&method=GET&min=2&max=5
```

Response:

```json
{
  "path": "/users",
  "method": "GET",
  "count": 3,
  "matched": true,
  "min": 2,
  "max": 5
}
```

Optional filters: `since` (timestamp), `status` (response status code).

**Implementation notes:**
- Reuse `config.RequestHistory` and existing `RequestRecord` fields
- Add query parsing in `pkg/server/server.go`
- Add tests in `pkg/server/server_test.go`
- Document in `docs/api_reference.md`

**Acceptance criteria:**
- Integration test can assert retry count without parsing full history manually

---

### Feature 2: Scenario state reset ✅

**Shipped:** `POST /api/control/reset-scenarios`

**Why:** Stateful sequences ("fail twice, then succeed") are not repeatable without reset.

**Proposed API:**

```
POST /api/control/reset-scenarios
POST /api/control/reset-scenarios?path=/users&method=GET
```

**Behavior:**
- Reset `scenario.Index` to 0 for matching scenarios
- Reset circuit breaker state to `closed` where applicable
- Return count of scenarios reset

**Implementation notes:**
- Add `config.ResetScenarios(path, method string) int`
- Iterate `scenarios` sync.Map, match path/method filter
- Handler in `pkg/server/server.go`

**Acceptance criteria:**
- Same test can run twice in CI without cross-test pollution

---

### Feature 3: Testcontainers-go example ✅

**Shipped:** `examples/testcontainers/`

**Why:** Go users need copy-paste CI integration. A comment in README is not enough.

**Deliverable:** `examples/testcontainers/main_test.go` (or `docs/examples/testcontainers_test.go`)

**Flow:**
1. Start container from `arun0009/flakymock:latest`
2. `POST /scenario` with retry sequence
3. Run client against mapped port
4. Assert status codes
5. `GET /api/verify/requests` (once Feature 1 ships) or `/history` interim
6. `POST /api/control/reset-scenarios` between subtests

**Acceptance criteria:**
- `go test ./examples/testcontainers/...` passes with Docker available
- README links to the example

---

## Phase 3 — Credibility for shared use (weeks 7–10)

### Feature 4: File-based persistence (WireMock-style)

**Why:** Teams want scenarios to survive restarts. Files are enough; no DB required initially.

**Layout:**

```
<root-dir>/
├── mappings/          # scenario JSON files
└── __files/           # response body files (optional, later)
```

**Config:**
- `SCENARIO_ROOT_DIR` env var (default: `.`)
- Load `mappings/*.json` and `mappings/*.yaml` on startup
- `POST /api/control/persist-scenarios` writes in-memory scenarios to disk

**Tenancy (later):** `mappings/<namespace>/` when needed — not required for v1.

### Feature 5: Scenario list API

```
GET /scenarios
```

Returns loaded scenarios (path, method, response count, current index). Helps debugging and tooling.

---

## Phase 4 — Distribution (ongoing)

| Action | Frequency | Goal |
|--------|-----------|------|
| Post comparison doc | Once, then link from README | SEO + "why not WireMock" |
| Add recipe per real API (Stripe, S3, etc.) | Monthly | Searchable use cases |
| Respond on "how to test retries" threads | As found | Point to repo + example |
| Short screen recording (30–60s) | Once | README / social embed |
| Ask for star only after user completes quickstart | In docs | Honest growth |

---

## Success metrics (realistic)

| Timeframe | Goal |
|-----------|------|
| 1 month after Phase 1 | 1 public post + updated README live |
| 2 months after Phase 2 | Verification + reset + testcontainers example shipped |
| 3 months | 25–50 stars, 1 external blog/issue mentioning the project |
| 6 months | 100+ stars, 2–3 real "used in our CI" mentions |

Stars follow traffic and proof, not feature count. **One great example beats five undocumented endpoints.**

---

## What not to do yet

- Multi-tenant DB backend (until file persistence + namespaces prove insufficient)
- Record/playback from live traffic (WireMock owns this story today)
- Claiming superiority over WireMock in README (undermines trust)
- Multi-protocol mocking scope (TCP, SMTP, etc. — different product)

---

## Immediate next actions (this week)

1. Read and share [comparison.md](comparison.md) when explaining the project
2. Walk through [resilience-testing-example.md](resilience-testing-example.md) locally
3. Open issues for: verification API, scenario reset, testcontainers example
4. Update GitHub repo description and topics
5. Draft one Show HN / blog post from the resilience example
