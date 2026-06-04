# Docker

Step 7 adds a reproducible Docker and Docker Compose baseline for the local FT12
stack. It is intended for development and smoke testing, not public internet
exposure.

## Quick Start

```powershell
docker compose up --build
```

Open:

- Web UI: `http://localhost:5173`
- API health: `http://localhost:8080/health`
- API readiness: `http://localhost:8080/api/v1/ready`

Stop and remove runtime volumes:

```powershell
docker compose down -v
```

## Services

| Service | Image source | Internal ports | Host ports |
| --- | --- | --- | --- |
| `ft12-emulator` | `deploy/docker/go-service.Dockerfile` | `9000`, `9100` | `9000`, `9100` |
| `ft12-gateway` | `deploy/docker/go-service.Dockerfile` | `9200` | `9200` |
| `ft12-api` | `deploy/docker/go-service.Dockerfile` | `8080` | `8080` |
| `ft12-web` | `web/Dockerfile` | `8080` | `5173` |
| `prometheus` | `prom/prometheus:v3.0.1` | `9090` | `9090` |

Prometheus is optional and runs only with the `observability` profile.

## Compose Flow

```text
ft12-emulator:9000  <- ft12-gateway
ft12-emulator:9100  <- ft12-api
ft12-gateway:9200   <- ft12-api
ft12-api:8080       <- ft12-web nginx proxy
```

The browser talks to `ft12-web` on `localhost:5173`. Nginx serves the static
Vite build and proxies `/api`, `/health`, and `/metrics` to `ft12-api:8080`.
This keeps the frontend bundle on relative URLs and avoids service-name DNS in
the browser.

## Healthchecks

Go service images include `/app/ft12-healthcheck`.

- Emulator healthcheck: TCP connect to `127.0.0.1:9100`.
- Gateway healthcheck: TCP connect to `127.0.0.1:9200`.
- API healthcheck: HTTP GET `http://127.0.0.1:8080/api/v1/ready`.
- Web healthcheck: nginx probes proxied `/health`.

## Hardening

- Go services use a multi-stage build and distroless Debian runtime.
- Go runtime containers run as `nonroot:nonroot`.
- The final Go image contains only the service binary and healthcheck helper.
- The web image builds with Node and serves static files from nginx as the
  `nginx` user on non-privileged port `8080`.
- No secrets, `.env`, `node_modules`, `web/dist`, logs, or binaries are included
  in the Docker build context.

## Commands

```powershell
docker compose config
docker compose build
docker compose up -d
docker compose ps
docker compose logs -f
docker compose down -v
```

Smoke:

```powershell
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/api/v1/ready
curl http://127.0.0.1:8080/api/v1/overview
curl http://127.0.0.1:8080/api/v1/events
curl http://127.0.0.1:8080/api/v1/export/events.json
curl http://127.0.0.1:8080/api/v1/export/events.csv
curl http://127.0.0.1:8080/api/v1/export/overview.json
curl http://127.0.0.1:8080/metrics
```

## Limitations

There is no authentication, TLS, persistence, secret management, Kubernetes, or
cloud deployment in this milestone. Do not expose this Compose stack to
untrusted networks.
