$ErrorActionPreference = "Stop"

$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$files = @()
$files += Get-Item (Join-Path $root "README.md")
$files += Get-Item (Join-Path $root "CONTRIBUTING.md") -ErrorAction SilentlyContinue
$files += Get-Item (Join-Path $root "SECURITY.md") -ErrorAction SilentlyContinue
$files += Get-Item (Join-Path $root "CHANGELOG.md") -ErrorAction SilentlyContinue
$files += Get-ChildItem -Path (Join-Path $root "docs") -Recurse -Filter "*.md" -ErrorAction SilentlyContinue
$files += Get-ChildItem -Path (Join-Path $root "examples") -Recurse -Filter "*.md" -ErrorAction SilentlyContinue
$files += Get-ChildItem -Path (Join-Path $root ".github") -Recurse -Filter "*.md" -ErrorAction SilentlyContinue

$pattern = '!?(\[[^\]]+\])\(([^)]+)\)'
$failures = New-Object System.Collections.Generic.List[string]

foreach ($file in $files | Where-Object { $_ }) {
    $content = Get-Content -Raw -Path $file.FullName
    $matches = [regex]::Matches($content, $pattern)
    foreach ($match in $matches) {
        $target = $match.Groups[2].Value.Trim()
        if ($target -eq "" -or $target.StartsWith("#")) {
            continue
        }
        if ($target -match '^(https?://|mailto:|tel:)') {
            continue
        }
        if ($target.StartsWith("<") -and $target.EndsWith(">")) {
            $target = $target.Substring(1, $target.Length - 2)
        }
        $target = ($target -split '\s+')[0]
        $target = ($target -split '#')[0]
        if ($target -eq "") {
            continue
        }
        $target = [System.Uri]::UnescapeDataString($target)
        if ([System.IO.Path]::IsPathRooted($target)) {
            $candidate = Join-Path $root $target.TrimStart("\", "/")
        } else {
            $candidate = Join-Path $file.DirectoryName $target
        }
        if (-not (Test-Path -LiteralPath $candidate)) {
            $relFile = [System.IO.Path]::GetRelativePath($root, $file.FullName)
            $failures.Add("$relFile -> $($match.Groups[2].Value)")
        }
    }
}

if ($failures.Count -gt 0) {
    Write-Error ("broken local markdown links:`n" + ($failures -join "`n"))
    exit 1
}

Write-Output "doc link check passed"
