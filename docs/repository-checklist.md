# Repository Checklist

Before a public demo or a release tag.

## Source hygiene

- `git status` is clean.
- No `bin/`, `dist/`, `tmp/`, `runtime/`, logs or local `.exe` files are
  tracked.
- `.gitignore` and the per-Dockerfile ignore files cover runtime artefacts.

## Checks

```sh
go run ./tools/checks release
```

That covers formatting, dependency boundaries, documentation links, `go vet`,
the tests, the build, Compose configuration and a cleanup dry-run. Separately:

```sh
make lint
pnpm --dir web lint
pnpm --dir web typecheck
pnpm --dir web test
go test -race ./...
go run ./tools/checks coverage --min 70
```

## Smoke

```sh
docker compose up -d --build
```

- `GET http://127.0.0.1:8080/health` returns `200`.
- `GET http://127.0.0.1:8080/api/v1/ready` returns `200`.
- `GET http://127.0.0.1:8080/api/v1/overview` returns `200`.
- `GET http://127.0.0.1:8080/api/v1/events` returns `200`.
- `GET http://127.0.0.1:8080/api/v1/gateway/fleet` returns `200`.
- `GET http://127.0.0.1:8080/api/v1/export/events.csv` returns `200`.
- `GET http://127.0.0.1:8080/metrics` returns `200`.
- `GET http://127.0.0.1:3000/en` returns `200`.
- `GET http://127.0.0.1:3000/upstream/api/v1/gateway/fleet` returns `200` —
  the console's own proxy, which is the check that `FT12_API_URL` is read at
  run time.
- `docker compose ps` shows four healthy containers.
- No container logs an error while starting.

```sh
docker compose down -v
```

## Documentation

- The README quick start works from a fresh clone.
- The docs index links every document.
- The screenshots in [the tour](tour.md) match the current interface; re-run
  [tools/docmedia](../tools/docmedia/README.md) if they do not.
- [CHANGELOG.md](../CHANGELOG.md) describes what actually changed.
- Troubleshooting covers the failures that have actually happened.

## Safety

- The README and the docs state plainly that there is no auth, TLS or
  persistence.
- No secrets are committed.
- No production-readiness claim is made beyond the local baseline that exists.
