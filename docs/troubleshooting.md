# Troubleshooting

## Go Toolchain

The module currently declares Go `1.26`.

Symptoms:

- `go: errors parsing go.mod`
- `go: unknown directive`
- CI works but local build fails

Actions:

- Install the Go version declared in `go.mod`.
- Run `go version`.
- Run `go env GOMOD` from the repository root.

## Node And npm

Use Node 22 or another modern LTS release.

If `npm ci` fails on Windows with `EPERM` under `web/node_modules`, stop any
running Vite/npm processes and remove `web/node_modules` before retrying:

```powershell
Get-CimInstance Win32_Process | Where-Object { $_.Name -like 'node*' } | Select-Object ProcessId,CommandLine
Remove-Item -Recurse -Force web\node_modules
cd web
npm ci
```

## Docker Daemon Unavailable

Symptoms:

- `Cannot connect to the Docker daemon`
- `docker compose config` works but build/up fails

Actions:

- Start Docker Desktop or the Docker service.
- Run `docker info`.
- If Docker is unavailable, still run Go, frontend, architecture, and doc-link checks.

## Docker Registry EOF Or Pull Errors

Transient registry/network errors can happen while pulling `golang`, `nginx`,
`node`, distroless, or Prometheus images.

Actions:

- Retry `docker compose build`.
- Run `docker pull` for the failing image.
- Check network/proxy/firewall settings.

## Ports Already In Use

Default ports:

- `9000`: emulator TCP
- `9100`: emulator gRPC
- `9200`: gateway gRPC
- `8080`: HTTP API
- `5173`: Web UI
- `9090`: Prometheus

Windows check:

```powershell
Get-NetTCPConnection -LocalPort 9000,9100,9200,8080,5173,9090 -ErrorAction SilentlyContinue
```

Stop stale Compose services:

```powershell
docker compose down -v
```

## API Readiness Fails

`/api/v1/ready` returns `503` when the API cannot reach emulator or gateway
gRPC upstreams.

Actions:

- Check `docker compose ps`.
- Check `docker compose logs ft12-emulator ft12-gateway ft12-api`.
- Verify API flags:
  - `-emulator-grpc ft12-emulator:9100` in Compose
  - `-gateway-grpc ft12-gateway:9200` in Compose

## Web UI Cannot Reach API

In local Vite development, `/api` proxies to `http://localhost:8080`.

In Docker Compose, nginx proxies `/api` to `http://ft12-api:8080`.

Actions:

- Check `http://localhost:5173/health`.
- Check `http://localhost:8080/health`.
- Inspect browser devtools network errors.
- Ensure `VITE_API_BASE_URL` is empty for same-origin proxy deployment.

## Prometheus Target Unavailable

Run:

```powershell
docker compose --profile observability up -d --build
```

Open `http://localhost:9090`, then inspect targets. The expected scrape target
is `ft12-api:8080/metrics`.

## Windows PowerShell Notes

`make` and `sh` may be unavailable on Windows. Use PowerShell equivalents:

```powershell
.\scripts\check-architecture.ps1
.\scripts\check-doc-links.ps1
.\scripts\release-check.ps1
```

Git may warn that LF will be replaced by CRLF. That warning is expected in some
Windows configurations and does not imply a content change by itself.

## Windows CI Go Formatting Fails

Symptoms:

- `Backend (windows-latest)` prints a long list of `.go` files.
- The list includes active Go files and possibly legacy reference files.

Actions:

- Run `go fmt ./...`.
- Run `.\scripts\check-go-format.ps1`.
- Check that `.gitattributes` is present.
- Do not include `legacy/` in active formatting/build/test fixes unless legacy
  reference code is intentionally being changed.

The CI format script normalizes CRLF/LF while comparing `gofmt` output, so a
failure means real formatting drift in active Go code.

## Logs Or Runtime Files Accumulate

Service logs default to `runtime/logs`. This folder is ignored by Git.

Preview cleanup:

```powershell
.\scripts\clean-runtime.ps1 -DryRun
```

```sh
bash scripts/clean-runtime.sh --dry-run
```

Then remove ignored runtime/build artifacts:

```powershell
.\scripts\clean-runtime.ps1
```
