# FT12 Web Dashboard

React + TypeScript + Vite dashboard for the FT12 HTTP API.

## Development

```powershell
npm install
npm run dev
```

The dev server listens on `http://localhost:5173` and proxies `/api` to
`http://localhost:8080`.

## Checks

```powershell
npm run typecheck
npm run build
npm run lint
```

## Runtime

Set `VITE_API_BASE_URL` when the HTTP API is not available through the Vite
proxy or same-origin path.
