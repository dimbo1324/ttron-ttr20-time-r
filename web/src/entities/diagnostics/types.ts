export type HealthStatus = {
  status: string;
  service: string;
  version: string;
  commit?: string;
  buildDate?: string;
};

export type ReadyStatus = {
  status: string;
  emulator: string;
  gateway: string;
};
