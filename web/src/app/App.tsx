import { useMemo, useState } from 'react';
import { appIcon as AppIcon, routes, type RouteID } from './routes';

export function App() {
  const [routeID, setRouteID] = useState<RouteID>('dashboard');
  const active = useMemo(() => routes.find((route) => route.id === routeID) ?? routes[0], [routeID]);
  const Page = active.component;

  return (
    <div className="min-h-screen text-zinc-100">
      <aside className="fixed inset-y-0 left-0 hidden w-64 border-r border-line bg-graphite/95 p-4 md:block">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg border border-signal/40 bg-signal/10 text-signal">
            <AppIcon size={20} />
          </div>
          <div>
            <div className="font-semibold">FT12 Control</div>
            <div className="text-xs text-zinc-500">HTTP dashboard</div>
          </div>
        </div>
        <nav className="mt-8 space-y-2">
          {routes.map((route) => {
            const Icon = route.icon;
            const selected = active.id === route.id;
            return (
              <button
                key={route.id}
                className={`flex w-full items-center gap-3 rounded-md border px-3 py-2 text-left text-sm transition ${selected ? 'border-signal/40 bg-signal/10 text-signal' : 'border-transparent text-zinc-400 hover:border-line hover:bg-white/5 hover:text-zinc-100'}`}
                onClick={() => setRouteID(route.id)}
              >
                <Icon size={16} />
                {route.label}
              </button>
            );
          })}
        </nav>
      </aside>
      <header className="sticky top-0 z-10 border-b border-line bg-graphite/95 p-3 md:hidden">
        <select className="w-full rounded-md border border-line bg-panel px-3 py-2" value={routeID} onChange={(event) => setRouteID(event.target.value as RouteID)}>
          {routes.map((route) => <option key={route.id} value={route.id}>{route.label}</option>)}
        </select>
      </header>
      <main className="p-4 md:ml-64 md:p-6">
        <Page />
      </main>
    </div>
  );
}
