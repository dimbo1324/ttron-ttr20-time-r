# syntax=docker/dockerfile:1.7

# Base images are pinned by digest, not by tag. A tag is a moving pointer:
# `golang:1.26-bookworm` today and next month are two different filesystems,
# so a build that "worked yesterday" can quietly produce a different image
# from an unchanged Dockerfile. The tag stays in the comment so a reader can
# still tell what it is.

# golang:1.26-bookworm
FROM golang@sha256:512690a5660563b57d37ecc31129e7f136e831db2aed24a1dbeb8ad7380dc0fa AS builder
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

# gcr.io/distroless/base-debian12:nonroot
FROM gcr.io/distroless/base-debian12@sha256:7f0c72cd138b442ae0deeb69c08b1acf5525439ba251a49ad93c320a061567e5
WORKDIR /app
COPY --from=builder /out/service /app/service
COPY --from=builder /out/ft12-healthcheck /app/ft12-healthcheck
USER nonroot:nonroot
ENTRYPOINT ["/app/service"]
