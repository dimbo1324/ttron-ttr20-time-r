import { apiRequest, apiTextRequest } from '../../shared/api/client';
import type { HealthStatus, ReadyStatus } from './types';

export function getHealth() {
  return apiRequest<HealthStatus>('/health');
}

export function getReady() {
  return apiRequest<ReadyStatus>('/api/v1/ready');
}

export function getMetricsText() {
  return apiTextRequest('/metrics');
}
