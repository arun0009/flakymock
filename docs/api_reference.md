<div align="center">
  <img src="flakymock.png" alt="FlakyMock Mascot" width="120"/>
</div>



# API Reference

`flakymock` provides several built-in endpoints for control, stress testing, and observability.

## Core Endpoints

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/echo` | `ANY` | Echoes the request body. Useful for testing request/response handling. |
| `/history` | `GET` | Returns a JSON array of the last `HISTORY_SIZE` requests. |

## Observability

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/health` | `GET` | Health check endpoint with uptime tracking, system info (goroutines, OS, arch), and extensible health checks. |
| `/metrics` | `GET` | Prometheus metrics endpoint. |

## Stress Testing (Chaos)

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/api/stress/cpu/{duration}` | `GET` | Consumes CPU for the specified duration (e.g., `/api/stress/cpu/5s`). |
| `/api/stress/mem/{size}` | `GET` | Allocates memory of specified size (e.g., `/api/stress/mem/100MB`). |

## Streaming

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/ws` | `GET` | Websocket echo endpoint. Connect and send messages to have them echoed back. |
| `/sse` | `GET` | Server-Sent Events endpoint. Streams the current time every 2 seconds. |

## Scenarios

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/scenario` | `POST` | Add scenario(s). Body: JSON object or array. Returns `ids`. |
| `/scenario` | `PUT` | Replace scenario by `id` (required in body). |
| `/scenarios` | `GET` | List loaded scenarios. Optional query: `path`, `method`. |
| `/scenarios` | `DELETE` | Delete by `id` or `path`+`method`. Removes mapping file by default. |

## Control

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/api/control/reset-history` | `POST` | Clears the request history. |
| `/api/control/reset-metrics` | `POST` | Resets the `mock_faults_injected_total` metric. |
| `/api/control/reset-scenarios` | `POST` | Resets scenario response indices and circuit breaker state. Optional query: `path`, `method`. |
| `/api/control/persist-scenarios` | `POST` | Write in-memory scenarios to `SCENARIO_ROOT_DIR/mappings/`. |
| `/replay` | `POST` | Replays a past request. Body: `{"id": "123", "target": "http://..."}`. |

## Verification

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/api/verify/requests` | `GET` | Count/filter recorded requests. Query: `path`, `method`, `status`, `since`, `min`, `max`, `body_contains`, `header=Name:value` (repeatable), `min_interval` (e.g. `100ms`), `detail=true`. |
| `/api/verify/sequence` | `POST` | Assert calls occurred in order. Body: `{"steps":[{"path":"/a","method":"GET"},{"path":"/b","method":"POST"}]}`. |
