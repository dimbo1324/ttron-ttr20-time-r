param(
    [switch] $DryRun
)

$ErrorActionPreference = "Stop"

$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$relativeTargets = @(
    "tmp",
    "runtime",
    "bin",
    "dist",
    "coverage.out",
    "web/dist",
    "web/.vite",
    "web/tsconfig.app.tsbuildinfo"
)
$filePatterns = @("*.log", "*.out", "*.tsbuildinfo")

function Assert-InRepo([string] $path) {
    $full = [System.IO.Path]::GetFullPath($path)
    if (-not $full.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "refusing to clean path outside repository: $full"
    }
    return $full
}

function Remove-SafePath([string] $path) {
    if (-not (Test-Path -LiteralPath $path)) {
        return
    }
    $full = Assert-InRepo $path
    $rel = $full.Substring($root.Length).TrimStart("\", "/")
    if ($DryRun) {
        Write-Output "would remove $rel"
        return
    }
    Write-Output "remove $rel"
    Remove-Item -LiteralPath $full -Recurse -Force
}

Push-Location $root
try {
    foreach ($target in $relativeTargets) {
        Remove-SafePath (Join-Path $root $target)
    }

    foreach ($pattern in $filePatterns) {
        Get-ChildItem -Path $root -Filter $pattern -File -ErrorAction SilentlyContinue |
            Where-Object { $_.FullName -notlike (Join-Path $root ".git*") } |
            ForEach-Object { Remove-SafePath $_.FullName }
    }
} finally {
    Pop-Location
}

if ($DryRun) {
    Write-Output "cleanup dry-run complete"
} else {
    Write-Output "cleanup complete"
}
