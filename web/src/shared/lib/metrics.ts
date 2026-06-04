export type MetricsSummary = {
  requestsTotal: number;
  paths: Array<{ path: string; count: number }>;
};

export function parseMetricsSummary(text: string): MetricsSummary {
  const counts = new Map<string, number>();
  for (const line of text.split('\n')) {
    if (!line.startsWith('ft12_http_requests_total')) continue;
    const parts = line.trim().split(/\s+/);
    const value = Number(parts[parts.length - 1]);
    if (!Number.isFinite(value)) continue;
    const path = /path="([^"]+)"/.exec(line)?.[1] ?? 'unknown';
    counts.set(path, (counts.get(path) ?? 0) + value);
  }
  const paths = [...counts.entries()]
    .map(([path, count]) => ({ path, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 6);
  return {
    requestsTotal: paths.reduce((sum, item) => sum + item.count, 0),
    paths,
  };
}
