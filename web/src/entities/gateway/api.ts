import { apiRequest } from '../../shared/api/client';
import type { GatewayStatus, LastReadTime } from './types';

export function getGatewayStatus() {
  return apiRequest<GatewayStatus>('/api/v1/gateway/status');
}

export function startGatewayPolling() {
  return apiRequest<{ status: GatewayStatus }>('/api/v1/gateway/start', { method: 'POST' });
}

export function stopGatewayPolling() {
  return apiRequest<{ status: GatewayStatus }>('/api/v1/gateway/stop', { method: 'POST' });
}

export function getLastReadTime() {
  return apiRequest<LastReadTime>('/api/v1/gateway/last-read-time');
}
