# FT12 Web Dashboard

React + TypeScript + Vite dashboard for the FT12 HTTP API.

The UI is Russian by default, can switch to English, supports persisted
dark/light themes, includes protocol infographics, and exposes JSON/CSV export
controls for local analysis.

## Development

```powershell
npm install
npm run dev
```

The dev server listens on `http://localhost:5173` and proxies `/api` to
`http://localhost:8080`.

## Features

- Compact responsive dashboard.
- RU/EN localisation stored in `localStorage`.
- Dark/light theme stored in `localStorage`.
- Protocol flow, frame anatomy, polling timeline, and event distribution widgets.
- Events JSON/CSV export plus overview and service status JSON export.
- Diagnostics and Guide pages.

## Checks

```powershell
npm run typecheck
npm run build
npm run lint
```

## Runtime

Set `VITE_API_BASE_URL` when the HTTP API is not available through the Vite
proxy or same-origin path.
