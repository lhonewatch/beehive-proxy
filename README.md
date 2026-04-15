# beehive-proxy

A lightweight reverse proxy with built-in request tracing and latency histograms exposed via Prometheus metrics.

---

## Installation

```bash
go install github.com/yourusername/beehive-proxy@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/beehive-proxy.git
cd beehive-proxy
go build -o beehive-proxy ./cmd/beehive-proxy
```

---

## Usage

Start the proxy by pointing it at an upstream target:

```bash
beehive-proxy --listen :8080 --target http://localhost:3000
```

Prometheus metrics are automatically exposed at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

### Example: Docker Compose

```yaml
services:
  proxy:
    image: yourusername/beehive-proxy:latest
    ports:
      - "8080:8080"
    environment:
      - TARGET=http://app:3000
      - LISTEN=:8080
```

### Available Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--listen` | `:8080` | Address to listen on |
| `--target` | _(required)_ | Upstream service URL |
| `--metrics-path` | `/metrics` | Prometheus metrics endpoint |
| `--trace` | `false` | Enable request tracing headers |

---

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `beehive_request_duration_seconds` | Histogram | Latency of proxied requests |
| `beehive_requests_total` | Counter | Total number of proxied requests |
| `beehive_errors_total` | Counter | Total number of proxy errors |

---

## License

[MIT](LICENSE)