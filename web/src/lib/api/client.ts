import type { z } from "zod";

import {
  emulatorStatusSchema,
  eventsSchema,
  faultModeSchema,
  fleetSchema,
  gatewayCommandSchema,
  gatewayStatusSchema,
  historySchema,
  setFaultModeSchema,
  type ApiEmulatorStatus,
  type ApiEvent,
  type ApiFaultMode,
  type ApiFleet,
  type ApiGatewayStatus,
  type ApiHistory,
} from "./schema";

/**
 * The client for the Go HTTP API.
 *
 * Requests go to `/upstream/...` on this app's own origin, which `next.config`
 * rewrites to the API. That indirection is not decoration: it keeps the API
 * off the public network in a deployment, removes CORS from the picture, and
 * means the browser only ever knows one origin. Nothing here should ever be
 * given an absolute URL.
 */

export const API_BASE = "/upstream/api/v1";

/** Code used when a failure carried no error envelope of its own. */
export const NO_ANSWER = "NO_ANSWER";

/**
 * A failed call, carrying enough to say something specific on screen.
 *
 * `status` is 0 when the request never got an answer at all, and the code is
 * NO_ANSWER when the reply carried no error envelope.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(message: string, status: number, code: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }

  /**
   * True when the API process could not be reached.
   *
   * Not simply `status === 0`. Requests go through this app's own rewrite, so
   * a dead API is not a network failure in the browser — Next answers the
   * proxied request itself, with a bare 5xx and no body. The API's own
   * failures always carry the `{error:{code,message}}` envelope, and that
   * envelope is what tells the two apart: an unreachable API reads as
   * "nothing is running", while a 503 from the API means the API is fine and
   * the gateway behind it is not.
   */
  get offline(): boolean {
    return this.status === 0 || (this.code === NO_ANSWER && this.status >= 500);
  }
}

/** The error envelope the API uses for every failure it can describe. */
interface ErrorEnvelope {
  error?: { code?: string; message?: string };
}

/**
 * The schema is a generic parameter rather than `z.ZodType<T>` because these
 * schemas carry defaults: their *input* is looser than their output, and
 * pinning one type variable to both ends makes every one of them fail to
 * assign. Inferring the output off the schema keeps the defaults working.
 */
async function request<S extends z.ZodTypeAny>(
  path: string,
  schema: S,
  init?: RequestInit,
): Promise<z.infer<S>> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      // The console shows what the gateway is doing *now*; a cached status is
      // worse than no status, because it looks current.
      cache: "no-store",
      ...init,
      headers: { Accept: "application/json", ...init?.headers },
    });
  } catch (cause) {
    throw new ApiError(cause instanceof Error ? cause.message : "network error", 0, NO_ANSWER);
  }

  const body: unknown = await response.json().catch(() => null);

  if (!response.ok) {
    const envelope = body as ErrorEnvelope | null;
    throw new ApiError(
      envelope?.error?.message ?? `HTTP ${response.status}`,
      response.status,
      envelope?.error?.code ?? NO_ANSWER,
    );
  }

  const parsed = schema.safeParse(body);
  if (!parsed.success) {
    throw new ApiError(parsed.error.issues[0]?.message ?? "unexpected response", response.status, "BAD_SHAPE");
  }
  return parsed.data;
}

function json(payload: unknown): RequestInit {
  return {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  };
}

export const api = {
  gatewayStatus: (signal?: AbortSignal): Promise<ApiGatewayStatus> =>
    request("/gateway/status", gatewayStatusSchema, { signal }),

  gatewayFleet: (signal?: AbortSignal): Promise<ApiFleet> =>
    request("/gateway/fleet", fleetSchema, { signal }),

  gatewayHistory: (signal?: AbortSignal): Promise<ApiHistory> =>
    request("/gateway/history", historySchema, { signal }),

  gatewayEvents: (limit: number, signal?: AbortSignal): Promise<ApiEvent[]> =>
    request(`/gateway/events?limit=${limit}`, eventsSchema, { signal }).then((body) => body.events),

  startPolling: (signal?: AbortSignal): Promise<ApiGatewayStatus> =>
    request("/gateway/start", gatewayCommandSchema, { method: "POST", signal }).then(
      (body) => body.status,
    ),

  stopPolling: (signal?: AbortSignal): Promise<ApiGatewayStatus> =>
    request("/gateway/stop", gatewayCommandSchema, { method: "POST", signal }).then(
      (body) => body.status,
    ),

  emulatorStatus: (signal?: AbortSignal): Promise<ApiEmulatorStatus> =>
    request("/emulator/status", emulatorStatusSchema, { signal }),

  faultMode: (signal?: AbortSignal): Promise<ApiFaultMode> =>
    request("/emulator/fault-mode", faultModeSchema, { signal }),

  setFaultMode: (fault: ApiFaultMode, signal?: AbortSignal): Promise<ApiFaultMode> =>
    request("/emulator/fault-mode", setFaultModeSchema, { ...json(fault), signal }).then(
      (body) => body.faultMode,
    ),
};
