import { buildApiUrl } from '../api/client';

function filenameFromDisposition(disposition: string | null, fallback: string) {
  if (!disposition) return fallback;
  const match = /filename="([^"]+)"/.exec(disposition);
  return match?.[1] ?? fallback;
}

export async function downloadEndpoint(path: string, fallbackFilename: string) {
  const response = await fetch(buildApiUrl(path));
  if (!response.ok) {
    throw new Error(response.statusText || 'export failed');
  }
  const blob = await response.blob();
  const filename = filenameFromDisposition(response.headers.get('Content-Disposition'), fallbackFilename);
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);
}

export async function copyJSON(value: unknown) {
  await navigator.clipboard.writeText(JSON.stringify(value, null, 2));
}
