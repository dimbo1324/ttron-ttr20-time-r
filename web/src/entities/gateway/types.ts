export type GatewayStatus = {
  state: string;
  targetAddr: string;
  checksumMode: string;
  pollingIntervalMs: number;
  requestTimeoutMs: number;
  connectTimeoutMs: number;
  connected: boolean;
  connectionAttempts: number;
  successfulReads: number;
  failedReads: number;
  reconnects: number;
  lastSuccessfulReadTime?: string;
  lastDeviceTime?: string;
  lastError?: string;
  lastTxTime?: string;
  lastRxTime?: string;
  recentFramesCount: number;
};

export type LastReadTime = {
  available: boolean;
  deviceTime?: string;
  readTime?: string;
};
