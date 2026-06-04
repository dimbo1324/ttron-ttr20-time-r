# Web UI

Step 6 adds a React, TypeScript, Vite, and Tailwind dashboard under `web/`.
The UI talks to the Go HTTP API with ordinary HTTP/JSON. It does not use
gRPC-Web.

## Pages

- Dashboard: combined emulator/gateway status, last device time, and recent events.
- Emulator: emulator counters and fault-mode controls.
- Gateway: gateway connection status and polling start/stop controls.
- Events / Frames: merged or source-filtered recent frame/event table.
- Settings: public API wiring and local dashboard preferences.

## Local Development

```powershell
cd web
npm install
npm run dev
```

Open `http://localhost:5173`.

Vite proxies `/api` and `/health` to `http://localhost:8080`.

Docker Compose serves the production Vite build from nginx. The nginx container
proxies `/api`, `/health`, and `/metrics` to `ft12-api:8080`, so the frontend
continues to use same-origin relative URLs.

## Environment

`VITE_API_BASE_URL` can point the dashboard at a specific API origin. When it is
empty, the Vite proxy or same-origin deployment path is used.

## Checks

```powershell
npm run typecheck
npm run build
npm run lint
```

## Limitations

The UI uses polling, not WebSocket or SSE. Advanced service configuration
remains CLI-based. There is no authentication, TLS, or persistence in this
milestone, and the local dashboard must not be exposed to untrusted networks.
