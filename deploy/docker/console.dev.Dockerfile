# syntax=docker/dockerfile:1.7

# The console, for development.
#
# Deliberately not the production image with a flag flipped. That image is
# distroless and carries a traced subset of the dependency tree with no
# compiler in it, which is exactly what makes it small and exactly what makes
# it useless for `next dev`. Two images that admit they are different beat one
# that pretends to be both.
#
# Runs as root, keeps the full dependency tree, and expects the source to
# arrive over a bind mount -- see docker-compose.dev.yml.

ARG NODE_VERSION=22
ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-bookworm AS health
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY proto ./proto
RUN CGO_ENABLED=0 go build -trimpath -o /out/ft12-healthcheck ./cmd/ft12-healthcheck

FROM node:${NODE_VERSION}-bookworm-slim
WORKDIR /src

RUN npm install -g pnpm@9

# The dependency layer is baked in rather than mounted: installing a tree of
# this size over a bind mount from a Windows host takes minutes every start.
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# A starting point for anything the compose file does not mount, so the image
# runs on its own if someone forgets the overlay.
COPY web/ ./

COPY --from=health /out/ft12-healthcheck /app/ft12-healthcheck

ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
EXPOSE 3000

CMD ["pnpm", "dev"]
