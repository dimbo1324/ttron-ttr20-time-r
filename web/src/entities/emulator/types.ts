export type FaultMode = {
  responseDelayMs: number;
  corruptChecksum: boolean;
  corruptChecksumProbability: number;
  fragmentResponse: boolean;
  fragmentProbability: number;
  fragmentDelayMs: number;
  noResponse: boolean;
  closeAfterRequest: boolean;
};

export type EmulatorStatus = {
  state: string;
  listenAddr: string;
  checksumMode: string;
  activeConnections: number;
  totalConnections: number;
  totalRequests: number;
  totalResponses: number;
  protocolErrors: number;
  lastError?: string;
  lastRequestTime?: string;
  lastResponseTime?: string;
  faultMode?: FaultMode;
  recentFramesCount: number;
};
