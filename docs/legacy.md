# Legacy

The repository keeps previous implementations as reference material.

## Python

`legacy/python` contains the original Python prototype. It uses a related but
not identical FT1.2-like framing style and is useful for historical comparison.

## Go

`legacy/go_sln` contains the previous separate Go client/server modules. The
current active implementation has been migrated into the root monorepo under
`cmd/` and `internal/`.

## Build Policy

Legacy code is not part of the normal root `go test ./...` or `go build ./...`
workflow. It should not be treated as active product code.
