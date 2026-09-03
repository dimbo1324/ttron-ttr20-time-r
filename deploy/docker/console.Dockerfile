# syntax=docker/dockerfile:1.7

# The web console.
#
# Three stages and two toolchains. The Go one exists for a single binary:
# the runtime image is distroless, which has no shell and no curl, so the
# health check has to be a program -- and the stack already has one that both
# speaks HTTP and dials TCP. Building it here rather than inventing a second
# check keeps one answer to "is this service up" across every container.

ARG NODE_VERSION=22
ARG GO_VERSION=1.26

# ---------------------------------------------------------------- deps
# Dependencies are their own layer, keyed on the manifest and the lockfile, so
# editing a component does not reinstall the tree.
FROM node:${NODE_VERSION}-bookworm-slim AS deps
WORKDIR /src

# Pinned to the major that wrote the lockfile: pnpm 10 rewrites a v9 lockfile,
# which would make the image install something the author never ran.
RUN npm install -g pnpm@9

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# ---------------------------------------------------------------- builder
FROM node:${NODE_VERSION}-bookworm-slim AS builder
WORKDIR /src

RUN npm install -g pnpm@9
COPY --from=deps /src/node_modules ./node_modules
COPY web/ ./

# Telemetry is off by default in CI images: a build should not phone anywhere.
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm build

# ---------------------------------------------------------------- health
FROM golang:${GO_VERSION}-bookworm AS health
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY proto ./proto
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ft12-healthcheck ./cmd/ft12-healthcheck

# ---------------------------------------------------------------- runtime
FROM gcr.io/distroless/nodejs${NODE_VERSION}-debian12:nonroot
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# `output: "standalone"` traces the modules the app actually imports, so this
# carries the server and its real dependencies rather than the whole tree.
COPY --from=builder /src/.next/standalone ./
COPY --from=builder /src/.next/static ./.next/static
COPY --from=health /out/ft12-healthcheck /app/ft12-healthcheck

EXPOSE 3000
USER nonroot:nonroot

# The distroless node image runs node as its entrypoint, so this is the script
# to hand it -- the server that `output: "standalone"` generated.
CMD ["server.js"]
