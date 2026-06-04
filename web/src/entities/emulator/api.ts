import { apiRequest } from '../../shared/api/client';
import type { EmulatorStatus, FaultMode } from './types';

export function getEmulatorStatus() {
  return apiRequest<EmulatorStatus>('/api/v1/emulator/status');
}

export function getFaultMode() {
  return apiRequest<FaultMode>('/api/v1/emulator/fault-mode');
}

export function updateFaultMode(faultMode: FaultMode) {
  return apiRequest<{ faultMode: FaultMode; status: EmulatorStatus }>('/api/v1/emulator/fault-mode', {
    method: 'PUT',
    body: JSON.stringify(faultMode),
  });
}
