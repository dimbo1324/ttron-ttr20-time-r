import { api, ApiError, API_BASE } from "./client";
import { faultModeFixture, gatewayStatusFixture, historyFixture } from "@/test/utils";

/**
 * The API client is the one place in this app that trusts nothing, so the
 * tests are mostly about what it does with bad news: a dead process, a
 * gateway error, a body that is not the shape it claims.
 */

/**
 * A stand-in for `Response`, which jsdom does not provide.
 *
 * Only the three members the client actually touches, which also keeps the
 * fixtures honest: a test cannot lean on a Response feature the client does
 * not use.
 */
function respond(body: unknown, init: { status?: number } = {}): Promise<Response> {
  const status = init.status ?? 200;
  return Promise.resolve({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  } as Response);
}

/** A reply whose body cannot be parsed as JSON at all. */
function respondUnparseable(status = 200): Promise<Response> {
  return Promise.resolve({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.reject(new SyntaxError("Unexpected end of JSON input")),
  } as Response);
}

function mockFetch(impl: (url: string, init?: RequestInit) => Promise<Response>) {
  const spy = jest.fn(impl);
  global.fetch = spy as unknown as typeof fetch;
  return spy;
}

describe("request routing", () => {
  it("goes through this app's own origin, never the API's", async () => {
    const fetchSpy = mockFetch(() => respond(gatewayStatusFixture()));

    await api.gatewayStatus();

    const [url] = fetchSpy.mock.calls[0]!;
    expect(url).toBe(`${API_BASE}/gateway/status`);
    // An absolute URL here would put the API on the public network and bring
    // CORS back; the rewrite exists precisely to avoid both.
    expect(url).not.toMatch(/^https?:/);
  });

  it("never serves a cached status", async () => {
    const fetchSpy = mockFetch(() => respond(gatewayStatusFixture()));

    await api.gatewayStatus();

    expect(fetchSpy.mock.calls[0]![1]).toMatchObject({ cache: "no-store" });
  });

  it("passes the event limit through", async () => {
    const fetchSpy = mockFetch(() => respond({ events: [] }));

    await api.gatewayEvents(25);

    expect(fetchSpy.mock.calls[0]![0]).toBe(`${API_BASE}/gateway/events?limit=25`);
  });

  it("starts and stops with a POST", async () => {
    const fetchSpy = mockFetch(() => respond({ status: gatewayStatusFixture() }));

    await api.startPolling();
    await api.stopPolling();

    expect(fetchSpy.mock.calls[0]![1]).toMatchObject({ method: "POST" });
    expect(fetchSpy.mock.calls[1]![0]).toBe(`${API_BASE}/gateway/stop`);
  });

  it("sends the fault mode as a JSON body", async () => {
    const fetchSpy = mockFetch(() => respond({ faultMode: faultModeFixture({ noResponse: true }) }));

    const got = await api.setFaultMode(faultModeFixture({ noResponse: true }));

    const init = fetchSpy.mock.calls[0]![1]!;
    expect(init.method).toBe("PUT");
    expect(JSON.parse(String(init.body))).toMatchObject({ noResponse: true });
    // The envelope also carries a status; only the fault mode is read back.
    expect(got.noResponse).toBe(true);
  });
});

describe("failures", () => {
  it("reports a process that never answered as offline", async () => {
    mockFetch(() => Promise.reject(new TypeError("Failed to fetch")));

    const error = await api.gatewayStatus().catch((cause: unknown) => cause);

    expect(error).toBeInstanceOf(ApiError);
    expect((error as ApiError).status).toBe(0);
    expect((error as ApiError).offline).toBe(true);
  });

  it("copes with a rejection that is not an Error", async () => {
    // Not every environment rejects fetch with an Error; the console still
    // has to say something rather than render "undefined".
    mockFetch(() => Promise.reject("connection refused"));

    const error = (await api.gatewayStatus().catch((cause: unknown) => cause)) as ApiError;

    expect(error.offline).toBe(true);
    expect(error.message).toBe("network error");
  });

  it("reports a bare 5xx as offline, because that is the proxy answering", async () => {
    // Requests go through this app's rewrite, so a dead API is not a network
    // failure in the browser -- Next answers with a 500 and no envelope.
    mockFetch(() => respond({}, { status: 500 }));

    const error = (await api.gatewayStatus().catch((cause: unknown) => cause)) as ApiError;

    expect(error.offline).toBe(true);
    expect(error.message).toBe("HTTP 500");
  });

  it("keeps an error the API chose to send, and does not call it offline", async () => {
    mockFetch(() =>
      respond(
        { error: { code: "GATEWAY_UNAVAILABLE", message: "gateway gRPC service is unavailable" } },
        { status: 503 },
      ),
    );

    const error = (await api.gatewayStatus().catch((cause: unknown) => cause)) as ApiError;

    // The API is up and telling us about the gateway behind it; saying "API
    // unreachable" here would send the reader after the wrong process.
    expect(error.offline).toBe(false);
    expect(error.code).toBe("GATEWAY_UNAVAILABLE");
    expect(error.message).toBe("gateway gRPC service is unavailable");
  });

  it("rejects a body that is not the shape it claims", async () => {
    mockFetch(() => respond({ clockSamples: "not a list" }));

    const error = (await api.gatewayHistory().catch((cause: unknown) => cause)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.code).toBe("BAD_SHAPE");
  });

  it("survives a success with no body at all", async () => {
    mockFetch(() => respondUnparseable());

    // A 200 with an unparseable body is not a shape this client can use, and
    // it fails at the boundary rather than four components later.
    await expect(api.gatewayStatus()).rejects.toBeInstanceOf(ApiError);
  });
});

describe("parsing", () => {
  it("fills in a field an older gateway does not send", async () => {
    const { protocolErrors, ...older } = gatewayStatusFixture();
    void protocolErrors;
    mockFetch(() => respond(older));

    const status = await api.gatewayStatus();

    // Degrading to a zero keeps the page up; taking it down over one field a
    // running binary predates would be the worse failure.
    expect(status.protocolErrors).toBe(0);
    expect(status.successfulReads).toBe(42);
  });

  it("fills in a whole nested section that is missing", async () => {
    const { clock, ...older } = gatewayStatusFixture();
    void clock;
    mockFetch(() => respond(older));

    const status = await api.gatewayStatus();

    expect(status.clock.state).toBe("unknown");
    expect(status.clock.samples).toBe(0);
  });

  it("reads the history windows", async () => {
    mockFetch(() => respond(historyFixture()));

    const history = await api.gatewayHistory();

    expect(history.clockSamples).toHaveLength(2);
    expect(history.healthOutcomes[1]!.success).toBe(false);
  });

  it("unwraps the events envelope", async () => {
    mockFetch(() => respond({ events: [{ id: 7, direction: "RX" }] }));

    const events = await api.gatewayEvents(10);

    expect(events).toHaveLength(1);
    expect(events[0]!.id).toBe(7);
  });

  it("reads the fleet, including a gateway that reports none", async () => {
    mockFetch(() => respond({}));

    const fleet = await api.gatewayFleet();

    expect(fleet.devices).toEqual([]);
    expect(fleet.summary.devices).toBe(0);
  });

  it("reads the emulator status and its fault mode", async () => {
    mockFetch((url) =>
      url.endsWith("/emulator/status")
        ? respond({ listenAddr: "127.0.0.1:9000", faultMode: faultModeFixture() })
        : respond(faultModeFixture({ responseDelayMs: 120 })),
    );

    expect((await api.emulatorStatus()).listenAddr).toBe("127.0.0.1:9000");
    expect((await api.faultMode()).responseDelayMs).toBe(120);
  });
});
