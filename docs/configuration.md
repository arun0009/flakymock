# Configuration

`flakymock` can be configured via `scenarios.yaml` for fault injection and environment variables for server settings.

## Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT` | Server port | `8080` |
| `ENABLE_TLS` | Enable HTTPS | `false` |
| `CERT_FILE` | TLS certificate path | `cert.pem` |
| `KEY_FILE` | TLS key path | `key.pem` |
| `ENABLE_CORS` | Enable CORS for all origins | `true` |
| `LOG_REQUESTS` | Log each request to stdout | `true` |
| `LOG_HEADERS` | Log request headers | `false` |
| `LOG_BODY` | Log request body | `true` |
| `MAX_BODY_SIZE` | Max request body size in bytes | `1048576` (1MB) |
| `RATE_LIMIT_RPS` | Requests per second limit (`0` = unlimited) | `0` |
| `HISTORY_SIZE` | Number of requests to keep in history | `100` |
| `ECHO_DELAY` | Global delay for all echo requests (e.g. `100ms`) | `0` |
| `ECHO_CHAOS_PROBABILITY` | Probability (0.0–1.0) of random 500 errors | `0.0` |
| `SCENARIO_ROOT_DIR` | Root for file persistence; loads `mappings/` on startup | `.` |

## Scenario Configuration (YAML)

Define multi-step response sequences in `scenarios.yaml`. See [Scenarios](scenarios.md) for full details.
