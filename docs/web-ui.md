# Web UI

The React, TypeScript, Vite, and Tailwind dashboard lives under `web/`. The UI
talks to the Go HTTP API with ordinary HTTP/JSON and does not use gRPC-Web.

## Pages

- Dashboard: compact emulator/gateway/API overview, status cards, protocol flow,
  event distribution, live process terminal, and recent events.
- Emulator: emulator counters, runtime status, frame anatomy, and fault-mode controls.
- Gateway: gateway counters, polling timeline, timing details, and start/stop controls.
- Events / Frames: source/direction filters, frame anatomy, event distribution,
  live process terminal, expandable raw hex, and JSON/CSV export controls.
- Diagnostics: API health/readiness, service counters, metrics summary, docs
  links, and export controls.
- Guide / Руководство: bilingual project guide and safety note.
- Settings: public API wiring plus local language/theme preferences.

## Localisation

Default language is Russian. The UI can switch between Russian and English with
the RU/EN control in the sidebar, mobile header, or Settings page. The selected
language is stored in browser `localStorage` under `ft12-ui-language`.

The lightweight i18n layer is intentionally dependency-free:

```text
web/src/shared/i18n/
  dictionaries/
    en.ts
    ru.ts
  I18nContext.ts
  i18n.tsx
  types.ts
  useI18n.ts
```

Missing translations fall back to English or the key name. Technical protocol
tokens such as `RX`, `TX`, `ERR`, `SYSTEM`, `SUM`, `CRC16`, `FT1.2`, and raw hex
remain technical unless a UI status label is intentionally mapped for display.

## Themes

Default theme is dark. Users can switch between dark and light themes; the
selection is stored in `localStorage` under `ft12-ui-theme`.

The theme layer uses CSS variables and Tailwind semantic tokens. Dark mode uses
graphite/slate surfaces with cyan and green accents. Light mode uses a clean
light background with the same industrial signal colors and stronger contrast
for tables, panels, forms, badges, and raw hex blocks.

Theme files:

```text
web/src/shared/theme/
  ThemeContext.ts
  ThemeProvider.tsx
  theme.ts
  useTheme.ts
```

## Layout And Interaction

The dashboard is denser than the first MVP UI. Sidebar width, card padding,
page spacing, and table height are reduced so the key status cards,
infographics, and recent events fit better on 1366x768, 1440x900, and
1920x1080 screens. Long tables use internal scrolling instead of growing the
whole page.

The second UI polish pass adds container-safe text wrapping and auto-fit
infographic grids. Buttons, badges, status cards, detail rows, table cells, and
protocol tiles allow long Russian/English labels, endpoint addresses, timestamps,
and errors to wrap instead of overlapping or being silently clipped. Shared
`InfoTile` and `DetailList` primitives keep this behavior consistent across
Dashboard, Emulator, Gateway, Events, Diagnostics, Settings, and Guide pages.

Interactive controls include hover/press feedback, disabled states, visible
keyboard focus, keyboard-expandable event rows, compact toggles, tooltips,
action result notices, and reduced motion support through
`prefers-reduced-motion`.

Main sections are deep-linkable with hash URLs such as `/#events`,
`/#emulator`, and `/#diagnostics`, so screenshots, bookmarks, and support links
can point users to the exact screen they should inspect.

## Live Terminal

Dashboard and Events include a live process terminal. It is a compact
operator-style feed built from the same recent event data as the tables:

- a running ticker summarizes the latest RX/TX/ERR/SYSTEM activity;
- the log area keeps recent frame/process lines in timestamp order;
- directions use the same signal colors as badges and charts;
- empty state text explains that the stack or gateway polling may need time to
  produce events.

The terminal is visual only; it does not open a shell and does not execute user
commands.

## Infographics

The UI includes lightweight CSS/SVG-free infographic components under
`web/src/widgets/infographics/`:

- `ProtocolFlow.tsx`: Web UI -> HTTP API -> gRPC -> Gateway -> FT1.2 TCP -> Emulator.
- `FrameAnatomy.tsx`: `0x68 | LEN | 0x68 | CONTROL | ADDRESS | DATA | CHECKSUM | 0x16`.
- `PollingTimeline.tsx`: connect, TX request, RX response, parse, status update,
  and retry/error indicators.
- `EventDistribution.tsx`: recent events by direction and by source.

## Exports

The UI exposes export and copy controls where the data is useful for analysis:

- Events JSON: `GET /api/v1/export/events.json?source=all&limit=100`
- Events CSV: `GET /api/v1/export/events.csv?source=all&limit=100`
- Overview JSON: `GET /api/v1/export/overview.json`
- Emulator status JSON: `GET /api/v1/export/emulator-status.json`
- Gateway status JSON: `GET /api/v1/export/gateway-status.json`

Downloaded filenames include a UTC timestamp, for example
`ft12-events-YYYYMMDD-HHMMSS.csv`. Copy JSON uses the browser clipboard and
renders data as text only.

Exports may contain protocol diagnostic data, raw hex, endpoint addresses, and
service counters. They are intended for local troubleshooting and analysis.

## Local Development

```powershell
cd web
npm ci
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
npm run lint
npm run build
```

## Manual QA Checklist

- RU language is the default.
- EN language switch updates navigation, pages, controls, empty/loading states,
  export controls, Guide, and infographics.
- Language persists after reload.
- Dark and light themes switch and persist after reload.
- Dashboard remains compact on 1366x768 and larger desktop viewports.
- Protocol flow and polling timeline tiles wrap cleanly without overlap when
  embedded in narrower dashboard columns.
- Long labels, status badges, endpoint addresses, timestamps, and errors remain
  readable without text clipping.
- Emulator fault toggles and numeric inputs update through the API.
- Gateway start/stop buttons show disabled/busy state during requests.
- Buttons show smooth press feedback, hover tooltips, and action result notices
  after polling, fault-mode, export, or copy actions.
- Events source/direction filters work.
- Live terminal renders recent events and its running ticker without overlapping
  surrounding content.
- Event rows expand with mouse, Enter, and Space.
- Raw hex is readable in both themes.
- JSON/CSV export buttons download files and do not fail on empty data.
- Copy JSON writes formatted JSON to the clipboard.
- Diagnostics health/readiness/metrics cards render without console errors.
- Guide page renders in both languages.

## Limitations

The UI uses polling, not WebSocket or SSE. Advanced service configuration
remains CLI-based. There is no authentication, TLS, or persistence in this
milestone, and the local dashboard must not be exposed to untrusted networks.
