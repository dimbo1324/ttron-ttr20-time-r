#!/usr/bin/env sh
set -eu

go fmt ./...
sh scripts/check-go-format.sh
sh scripts/check-architecture.sh
go test ./...
go build ./...

docker compose config >/dev/null
docker compose --profile observability config >/dev/null
sh scripts/check-doc-links.sh
bash scripts/clean-runtime.sh --dry-run

echo "release check passed"
