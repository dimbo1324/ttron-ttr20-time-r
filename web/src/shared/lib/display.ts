import type { EventSource } from '../../entities/events/types';
import type { TranslationKey } from '../i18n/types';

type Translate = (key: TranslationKey) => string;

const statusKeys: Record<string, TranslationKey> = {
  running: 'status.running',
  stopped: 'status.stopped',
  connected: 'status.connected',
  degraded: 'status.degraded',
  fault: 'status.fault',
  err: 'status.err',
  error: 'status.err',
  ready: 'status.ready',
  not_ready: 'status.notReady',
  ok: 'status.ok',
  unavailable: 'status.unavailable',
  unspecified: 'status.unspecified',
};

const sourceKeys: Record<EventSource, TranslationKey> = {
  all: 'source.all',
  emulator: 'source.emulator',
  gateway: 'source.gateway',
};

export function displayStatus(value: string | undefined, t: Translate): string {
  if (!value) return t('status.unspecified');
  return t(statusKeys[value.toLowerCase()] ?? 'status.unspecified');
}

export function displaySource(value: EventSource | string | undefined, t: Translate): string {
  if (!value) return t('status.unspecified');
  if (value === 'all' || value === 'emulator' || value === 'gateway') {
    return t(sourceKeys[value]);
  }
  return value;
}

export function displayChecksum(value: string | undefined): string {
  if (!value || value === 'unspecified') return 'unspecified';
  return value.toUpperCase();
}

export function statTone(value: string | undefined): 'default' | 'ok' | 'warn' | 'fault' | 'signal' {
  switch ((value ?? '').toLowerCase()) {
    case 'running':
    case 'connected':
    case 'ready':
    case 'ok':
      return 'ok';
    case 'degraded':
      return 'warn';
    case 'fault':
    case 'err':
    case 'error':
      return 'fault';
    default:
      return 'default';
  }
}
