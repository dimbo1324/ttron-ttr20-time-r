$ErrorActionPreference = "Stop"

function Fail($Message) {
    Write-Error "architecture check failed: $Message"
    exit 1
}

function Test-NoImport($Path, $Pattern, $Message) {
    $matches = Get-ChildItem -Path $Path -Recurse -Filter *.go -ErrorAction SilentlyContinue |
        Select-String -Pattern $Pattern -SimpleMatch
    if ($matches) {
        Fail $Message
    }
}

Test-NoImport "internal/protocol" '"net"' "internal/protocol must not import net"
Test-NoImport "internal/protocol" "google.golang.org/grpc" "internal/protocol must not import grpc"
Test-NoImport "internal/protocol" '"net/http"' "internal/protocol must not import http"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/config" "internal/protocol must not import internal/config"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/logging" "internal/protocol must not import internal/logging"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/logging" "internal/protocol must not import internal/platform/logging"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator" "internal/protocol must not import internal/emulator"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway" "internal/protocol must not import internal/gateway"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/transport" "internal/protocol must not import internal/transport"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/api" "internal/protocol must not import internal/api"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/app" "internal/protocol must not import internal/app"
Test-NoImport "internal/protocol" "github.com/dimbo1324/ttron-ttr20-time-r/internal/adapters" "internal/protocol must not import internal/adapters"
Test-NoImport "internal/emulator" "github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway" "internal/emulator must not import internal/gateway"
Test-NoImport "internal/gateway" "github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator" "internal/gateway must not import internal/emulator"

$legacyImports = Get-ChildItem -Path cmd,internal,proto -Recurse -Filter *.go -ErrorAction SilentlyContinue |
    Select-String -Pattern "github.com/dimbo1324/ttron-ttr20-time-r/legacy" -SimpleMatch
if ($legacyImports) {
    Fail "active code must not import legacy/"
}

$webImports = Get-ChildItem -Path cmd,internal,proto -Recurse -Filter *.go -ErrorAction SilentlyContinue |
    Select-String -Pattern "github.com/dimbo1324/ttron-ttr20-time-r/web" -SimpleMatch
if ($webImports) {
    Fail "Go packages must not import web source"
}

Get-ChildItem -Path "internal/api/grpc/ft12/v1" -Filter "*.pb.go" | ForEach-Object {
    if (-not (Select-String -Path $_.FullName -Pattern "Code generated" -SimpleMatch -Quiet)) {
        Fail "$($_.FullName) must contain Code generated marker"
    }
}

Write-Output "architecture check passed"
