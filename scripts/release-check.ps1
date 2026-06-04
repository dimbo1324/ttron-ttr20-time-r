$ErrorActionPreference = "Stop"

$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $root
try {
    foreach ($path in @("web\node_modules", "web\dist", "web\.vite", "web\tsconfig.app.tsbuildinfo")) {
        $resolved = Resolve-Path -LiteralPath $path -ErrorAction SilentlyContinue
        if ($resolved) {
            Remove-Item -LiteralPath $resolved.Path -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
    go fmt ./...
    .\scripts\check-go-format.ps1
    .\scripts\check-architecture.ps1
    go test ./...
    go build ./...
    Push-Location (Join-Path $root "web")
    try {
        npm ci
        npm run typecheck
        npm run lint
        npm run build
    } finally {
        Pop-Location
    }
    docker compose config | Out-Null
    docker compose --profile observability config | Out-Null
    .\scripts\check-doc-links.ps1
    .\scripts\clean-runtime.ps1 -DryRun
    Write-Output "release check passed"
} finally {
    Pop-Location
}
