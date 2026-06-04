# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

ARG SERVICE=ft12-api
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY proto ./proto

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X github.com/dimbo1324/ttron-ttr20-time-r/internal/version.Version=${VERSION} -X github.com/dimbo1324/ttron-ttr20-time-r/internal/version.Commit=${COMMIT} -X github.com/dimbo1324/ttron-ttr20-time-r/internal/version.BuildDate=${BUILD_DATE}" -o /out/service ./cmd/${SERVICE}
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ft12-healthcheck ./cmd/ft12-healthcheck

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/service /app/service
COPY --from=builder /out/ft12-healthcheck /app/ft12-healthcheck
USER nonroot:nonroot
ENTRYPOINT ["/app/service"]
