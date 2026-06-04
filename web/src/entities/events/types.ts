export type EventDirection = 'RX' | 'TX' | 'ERR' | 'SYSTEM';
export type EventSource = 'all' | 'emulator' | 'gateway';

export type FrameEvent = {
  id: number;
  timestamp?: string;
  source: string;
  service: string;
  direction: EventDirection;
  remoteAddr?: string;
  checksumMode: string;
  rawHex?: string;
  command?: string;
  error?: string;
  message?: string;
};

export type EventsResponse = {
  events: FrameEvent[];
  note?: string;
};

export type Overview = {
  health: {
    status: string;
    service: string;
    version: string;
  };
  emulator: import('../emulator/types').EmulatorStatus;
  gateway: import('../gateway/types').GatewayStatus;
  lastRead: import('../gateway/types').LastReadTime;
  events: FrameEvent[];
  eventsNote?: string;
};

export type PublicConfig = {
  emulatorGrpc: string;
  gatewayGrpc: string;
  pollingNote: string;
};
