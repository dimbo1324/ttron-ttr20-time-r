# Roadmap

1. Go monorepo baseline. Implemented.
2. FT1.2 protocol core. Implemented.
3. Emulator TCP service. Implemented.
4. Gateway polling service. Implemented.
5. gRPC contracts. Implemented.
5.5. Architecture hardening before Web UI. Implemented.
6. Web UI. Next.
7. Docker, observability, and CI.
8. Final docs and release polish.

Step 5 is complete when proto contracts, generated Go code, emulator/gateway
gRPC adapters, client helpers, integration tests, docs, and build/test checks
are in place.

Step 5.5 is complete when command entrypoints are thin, app bootstrap packages
own process lifecycle, config loading is testable and validated, event IDs are
stable, gRPC mapper duplication is reduced, service files are decomposed, and
architecture checks are available.
