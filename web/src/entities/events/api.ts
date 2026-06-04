import { apiRequest } from '../../shared/api/client';
import type { EventSource, EventsResponse, Overview, PublicConfig } from './types';

export function getEvents(source: EventSource = 'all', limit = 100) {
  return apiRequest<EventsResponse>(`/api/v1/events?source=${source}&limit=${limit}`);
}

export function getOverview() {
  return apiRequest<Overview>('/api/v1/overview');
}

export function getPublicConfig() {
  return apiRequest<PublicConfig>('/api/v1/config');
}
