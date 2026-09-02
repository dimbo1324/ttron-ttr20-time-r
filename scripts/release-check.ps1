$ErrorActionPreference = "Stop"

$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $root
try {
    go fmt ./...
    .\scripts\check-go-format.ps1
    .\scripts\check-architecture.ps1
    go test ./...
    go build ./...
    docker compose config | Out-Null
    docker compose --profile observability config | Out-Null
    .\scripts\check-doc-links.ps1
    .\scripts\clean-runtime.ps1 -DryRun
    Write-Output "release check passed"
} finally {
    Pop-Location
}
