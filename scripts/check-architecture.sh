#!/usr/bin/env sh
set -eu

fail() {
  echo "architecture check failed: $1" >&2
  exit 1
}

check_no_import() {
  path="$1"
  pattern="$2"
  message="$3"
  if grep -R --include='*.go' -n "$pattern" "$path" >/dev/null 2>&1; then
    fail "$message"
  fi
}

check_no_import internal/protocol '"net"' 'internal/protocol must not import net'
check_no_import internal/protocol 'google.golang.org/grpc' 'internal/protocol must not import grpc'
check_no_import internal/protocol 'github.com/dimbo1324/ttron-ttr20-time-r/internal/config' 'internal/protocol must not import internal/config'
check_no_import internal/protocol 'github.com/dimbo1324/ttron-ttr20-time-r/internal/logging' 'internal/protocol must not import internal/logging'
check_no_import internal/protocol 'github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/logging' 'internal/protocol must not import internal/platform/logging'
check_no_import internal/protocol 'github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator' 'internal/protocol must not import internal/emulator'
check_no_import internal/protocol 'github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway' 'internal/protocol must not import internal/gateway'
check_no_import internal/protocol 'github.com/dimbo1324/ttron-ttr20-time-r/internal/transport' 'internal/protocol must not import internal/transport'
check_no_import internal/protocol 'github.com/dimbo1324/ttron-ttr20-time-r/internal/api' 'internal/protocol must not import internal/api'
check_no_import internal/protocol 'github.com/dimbo1324/ttron-ttr20-time-r/internal/adapters' 'internal/protocol must not import internal/adapters'
check_no_import internal/emulator 'github.com/dimbo1324/ttron-ttr20-time-r/internal/gateway' 'internal/emulator must not import internal/gateway'
check_no_import internal/gateway 'github.com/dimbo1324/ttron-ttr20-time-r/internal/emulator' 'internal/gateway must not import internal/emulator'

if grep -R --include='*.go' -n 'github.com/dimbo1324/ttron-ttr20-time-r/legacy' cmd internal proto >/dev/null 2>&1; then
  fail 'active code must not import legacy/'
fi

for file in internal/api/grpc/ft12/v1/*.pb.go; do
  if ! grep -q 'Code generated' "$file"; then
    fail "$file must contain Code generated marker"
  fi
done

echo "architecture check passed"
