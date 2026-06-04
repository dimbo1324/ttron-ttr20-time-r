$ErrorActionPreference = "Stop"

$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$failures = New-Object System.Collections.Generic.List[string]

function Convert-ToLf([string] $value) {
    return ($value -replace "`r`n", "`n") -replace "`r", "`n"
}

Push-Location $root
try {
    $files = git ls-files "*.go" |
        Where-Object {
            $_ -and
            -not $_.StartsWith("legacy/") -and
            -not $_.StartsWith("web/node_modules/")
        }

    foreach ($file in $files) {
        $sourcePath = Join-Path $root $file
        $tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("ft12-gofmt-" + [System.Guid]::NewGuid().ToString("N"))
        New-Item -ItemType Directory -Path $tempDir | Out-Null
        $tempPath = Join-Path $tempDir ([System.IO.Path]::GetFileName($file))
        try {
            Copy-Item -LiteralPath $sourcePath -Destination $tempPath
            gofmt -w $tempPath

            $original = Convert-ToLf ([System.IO.File]::ReadAllText($sourcePath))
            $formatted = Convert-ToLf ([System.IO.File]::ReadAllText($tempPath))
            if ($original -ne $formatted) {
                $failures.Add($file)
            }
        } finally {
            Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
} finally {
    Pop-Location
}

if ($failures.Count -gt 0) {
    Write-Error ("Go files need gofmt:`n" + ($failures -join "`n"))
    exit 1
}

Write-Output "go format check passed"
