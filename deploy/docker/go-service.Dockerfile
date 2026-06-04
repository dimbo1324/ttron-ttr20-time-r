# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

ARG SERVICE=ft12-api
ARG VERSION=dev
ARG COMMIT=unknown

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY proto ./proto

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /out/service ./cmd/${SERVICE}
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ft12-healthcheck ./cmd/ft12-healthcheck

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/service /app/service
COPY --from=builder /out/ft12-healthcheck /app/ft12-healthcheck
USER nonroot:nonroot
ENTRYPOINT ["/app/service"]
