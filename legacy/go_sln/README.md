# Legacy Go Modules

This directory preserves the historical separate Go client/server modules.

The active implementation was migrated to the root Go monorepo:

- `cmd/ft12-client`
- `cmd/ft12-emulator`
- shared packages under `internal/`

The legacy modules are retained for comparison while the protocol platform is
modernized.
