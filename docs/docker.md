# Docker

The whole stack -- the device emulator, the polling gateway, the HTTP API and
the web console -- builds and runs from this repository with one command. It is
meant for development and smoke testing, not for exposure to the internet.

## Quick Start

```sh
docker compose up --build
```

Open:

- Console: `http://localhost:3000/ru` (or `/en`)
- API health: `http://localhost:8080/health`
- API readiness: `http://localhost:8080/api/v1/ready`

Stop and remove runtime volumes:

```sh
docker compose down -v
```

## Services

| Service | Image source | Internal ports | Host ports |
| --- | --- | --- | --- |
| `ft12-emulator` | `deploy/docker/go-service.Dockerfile` | `9000`, `9100` | `9000`, `9100` |
| `ft12-gateway` | `deploy/docker/go-service.Dockerfile` | `9200` | `9200` |
| `ft12-api` | `deploy/docker/go-service.Dockerfile` | `8080` | `8080` |
| `ft12-console` | `deploy/docker/console.Dockerfile` | `3000` | `3000` |
| `prometheus` | `prom/prometheus:v3.0.1` | `9090` | `9090` |

Prometheus is optional and runs only with the `observability` profile.

## Compose Flow

```text
ft12-emulator:9000  <- ft12-gateway
ft12-emulator:9100  <- ft12-api
ft12-gateway:9200   <- ft12-api
ft12-api:8080       <- ft12-console, HTTP clients
ft12-console:3000   <- browsers
```

## The Console And Its Proxy

The browser only ever talks to the console's own origin. The console forwards
`/upstream/*` to the API, which keeps the API off the public network, removes
CORS from the picture, and means the page never needs to know where the API is.

That forwarding is a route handler (`web/src/app/upstream`), not a
`rewrites()` entry, and the difference matters here specifically. Next resolves
a rewrite destination at **build** time and writes it into
`routes-manifest.json`, so an image built with the default would point at
`http://127.0.0.1:8080` forever -- inside a container, itself. `FT12_API_URL`
would be read during `docker build` and ignored at `docker run`, which is the
opposite of what an image is for. The handler reads it per request, so one
image runs anywhere.

## Logs

Every service is started with `-log ""`, which sends its log to stdout where
`docker logs` can see it.

This is stated rather than left to the default on purpose. The default is
`runtime/logs/<service>.log`, and inside a container that is `/app/runtime/logs`
with the process running as `nonroot` against a `/app` owned by root:

```text
cannot create log directory for runtime/logs/ft12-gateway.log:
mkdir runtime/logs: permission denied - logging to stdout
```

The fallback lands in the right place, but every container used to start with
an error about a directory nobody wanted.

## Development

```sh
docker compose -f docker-compose.yml -f docker-compose.dev.yml up
```

The overlay swaps the console for `next dev` against a bind-mounted source
tree, so an edit is on screen in a second rather than after an image rebuild,
and drops the gateway to a one-second interval because a bench nobody watches
for a minute teaches nothing. Everything else keeps its production wiring.

It is a named file rather than `docker-compose.override.yml`, which Compose
would pick up automatically: a stack that quietly runs in development mode
because a file happens to be on disk is how a demo ends up served by a dev
server that nobody noticed.

## Healthchecks

Every image carries `/app/ft12-healthcheck` -- including the console, whose
runtime image is distroless and has neither a shell nor curl. One answer to
"is this service up" across the stack beats four.

- Emulator: TCP connect to `127.0.0.1:9100`.
- Gateway: TCP connect to `127.0.0.1:9200`.
- API: HTTP GET `http://127.0.0.1:8080/api/v1/ready`.
- Console: HTTP GET `http://127.0.0.1:3000/ru`.

## Build Contexts

Each Dockerfile has its own ignore file next to it --
`go-service.Dockerfile.dockerignore`, `console.Dockerfile.dockerignore` --
which BuildKit prefers over the repository-wide `.dockerignore`. A Go service
copies `cmd/`, `internal/` and `proto/`; sending it the console's source as
well was the larger half of the context for no reason.

## Hardening

- Multi-stage builds throughout; the runtime images carry no toolchain.
- Both runtimes are distroless Debian and run as `nonroot:nonroot`.
- A Go image contains the service binary and the health check, nothing else.
- The console image carries the traced standalone output rather than the whole
  dependency tree, which is the difference between roughly 300 MB and the
  gigabyte a naive `node_modules` copy would cost.
- No secrets, `.env`, logs or binaries are included
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
