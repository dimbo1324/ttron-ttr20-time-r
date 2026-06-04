#!/usr/bin/env sh
set -eu

rm -rf web/node_modules web/dist web/.vite web/*.tsbuildinfo

go fmt ./...
sh scripts/check-architecture.sh
go test ./...
go build ./...

cd web
npm ci
npm run typecheck
npm run lint
npm run build
cd ..

docker compose config >/dev/null
docker compose --profile observability config >/dev/null
sh scripts/check-doc-links.sh

echo "release check passed"
