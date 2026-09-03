# syntax=docker/dockerfile:1.7

# The web console.
#
# Three stages and two toolchains. The Go one exists for a single binary:
# the runtime image is distroless, which has no shell and no curl, so the
# health check has to be a program -- and the stack already has one that both
# speaks HTTP and dials TCP. Building it here rather than inventing a second
# check keeps one answer to "is this service up" across every container.

# Base images are pinned by digest rather than by tag, so an unchanged
# Dockerfile cannot quietly produce a different image next month. The tag is
# kept in a comment above each.

# ---------------------------------------------------------------- deps
# Dependencies are their own layer, keyed on the manifest and the lockfile, so
# editing a component does not reinstall the tree.
# node:22-bookworm-slim
FROM node@sha256:83f487e0a63425e5b4d146fb5e5be574bcbe1b7b843d3ebafdd95eaf7767a7e5 AS deps
WORKDIR /src

# Pinned to the major that wrote the lockfile: pnpm 10 rewrites a v9 lockfile,
# which would make the image install something the author never ran.
RUN npm install -g pnpm@9

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# ---------------------------------------------------------------- builder
# node:22-bookworm-slim
FROM node@sha256:83f487e0a63425e5b4d146fb5e5be574bcbe1b7b843d3ebafdd95eaf7767a7e5 AS builder
WORKDIR /src

RUN npm install -g pnpm@9
COPY --from=deps /src/node_modules ./node_modules
COPY web/ ./

# Telemetry is off by default in CI images: a build should not phone anywhere.
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm build

# ---------------------------------------------------------------- health
# golang:1.26-bookworm
FROM golang@sha256:9fdc884aacc3bec89b20ffc69f4bb369c78210e3e4f600387b5128b12c199f81 AS health
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY proto ./proto
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ft12-healthcheck ./cmd/ft12-healthcheck

# ---------------------------------------------------------------- runtime
# gcr.io/distroless/nodejs22-debian12:nonroot
FROM gcr.io/distroless/nodejs22-debian12@sha256:13593b7570658e8477de39e2f4a1dd25db2f836d68a0ba771251572d23bb4f8e
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
