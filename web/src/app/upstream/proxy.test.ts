/**
 * @jest-environment node
 */
import { NextRequest } from "next/server";

import { GET, POST, PUT } from "./[...path]/route";

/**
 * The proxy between the browser and the Go API.
 *
 * It exists because a rewrite could not do the job: Next resolves a rewrite
 * destination at build time, so a container image would carry the address it
 * was built with rather than the one it was started with. Everything below is
 * about that difference, or about the contract the API client depends on --
 * that a failure with no error envelope means "nothing answered".
 */

const context = (...path: string[]) => ({ params: Promise.resolve({ path }) });

function request(url: string, init?: RequestInit): NextRequest {
  return new NextRequest(new Request(url, init));
}

function ok(body: unknown, init: { status?: number; headers?: Record<string, string> } = {}) {
  return Promise.resolve(
    new Response(JSON.stringify(body), {
      status: init.status ?? 200,
      headers: { "Content-Type": "application/json", ...init.headers },
    }),
  );
}

let fetchSpy: jest.Mock;

beforeEach(() => {
  delete process.env.FT12_API_URL;
  fetchSpy = jest.fn(() => ok({ state: "running" }));
  global.fetch = fetchSpy as unknown as typeof fetch;
});

describe("where it forwards to", () => {
  it("defaults to the API on this machine", async () => {
    await GET(request("http://console/upstream/api/v1/gateway/status"), context("api", "v1", "gateway", "status"));

    expect(fetchSpy.mock.calls[0]![0]).toBe("http://127.0.0.1:8080/api/v1/gateway/status");
  });

  it("reads the address at request time, not at build time", async () => {
    // The whole reason this is a handler. A rewrite would have baked the
    // default into routes-manifest.json, and a container told where the API
    // lives would have been ignored.
    process.env.FT12_API_URL = "http://ft12-api:8080";

    await GET(request("http://console/upstream/api/v1/health"), context("api", "v1", "health"));

    expect(fetchSpy.mock.calls[0]![0]).toBe("http://ft12-api:8080/api/v1/health");
  });

  it("tolerates a trailing slash on the configured address", async () => {
    process.env.FT12_API_URL = "http://ft12-api:8080/";

    await GET(request("http://console/upstream/api/v1/health"), context("api", "v1", "health"));

    expect(fetchSpy.mock.calls[0]![0]).toBe("http://ft12-api:8080/api/v1/health");
  });

  it("keeps the query string", async () => {
    await GET(
      request("http://console/upstream/api/v1/gateway/events?limit=200"),
      context("api", "v1", "gateway", "events"),
    );

    expect(fetchSpy.mock.calls[0]![0]).toBe("http://127.0.0.1:8080/api/v1/gateway/events?limit=200");
  });

  it("never serves a cached answer", async () => {
    await GET(request("http://console/upstream/api/v1/gateway/status"), context("api", "v1", "gateway", "status"));

    expect(fetchSpy.mock.calls[0]![1]).toMatchObject({ cache: "no-store" });
  });
});

describe("what it carries", () => {
  it("passes a POST through with its method", async () => {
    await POST(
      request("http://console/upstream/api/v1/gateway/start", { method: "POST" }),
      context("api", "v1", "gateway", "start"),
    );

    expect(fetchSpy.mock.calls[0]![1]).toMatchObject({ method: "POST" });
  });

  it("passes a PUT through with its body", async () => {
    await PUT(
      request("http://console/upstream/api/v1/gateway/settings", {
        method: "PUT",
        body: JSON.stringify({ pollIntervalMs: 2000 }),
        headers: { "Content-Type": "application/json" },
      }),
      context("api", "v1", "gateway", "settings"),
    );

    const init = fetchSpy.mock.calls[0]![1]!;
    expect(init.method).toBe("PUT");
    expect(JSON.parse(String(init.body))).toEqual({ pollIntervalMs: 2000 });
  });

  it("sends no body on a GET", async () => {
    // fetch throws outright when a GET is given one, so this is the
    // difference between a working proxy and one that fails every read.
    await GET(request("http://console/upstream/api/v1/health"), context("api", "v1", "health"));

    expect(fetchSpy.mock.calls[0]![1]!.body).toBeUndefined();
  });

  it("drops the headers that describe the hop rather than the message", async () => {
    await GET(
      request("http://console/upstream/api/v1/health", {
        headers: { Accept: "application/json", Connection: "keep-alive" },
      }),
      context("api", "v1", "health"),
    );

    const headers = fetchSpy.mock.calls[0]![1]!.headers as Headers;
    expect(headers.get("accept")).toBe("application/json");
    // `host` would name this app to a service that does not answer to it.
    expect(headers.get("host")).toBeNull();
    expect(headers.get("connection")).toBeNull();
  });

  it("returns the upstream status and body unchanged", async () => {
    fetchSpy.mockReturnValue(ok({ error: { code: "GATEWAY_UNAVAILABLE", message: "down" } }, { status: 503 }));

    const response = await GET(request("http://console/upstream/api/v1/gateway/status"), context("api", "v1", "gateway", "status"));

    expect(response.status).toBe(503);
    // The envelope has to survive: the client tells an API that refused
    // something from an API that is not there by whether it is present.
    await expect(response.json()).resolves.toMatchObject({ error: { code: "GATEWAY_UNAVAILABLE" } });
  });
});

describe("when nothing answers", () => {
  it("replies 502 with no envelope", async () => {
    fetchSpy.mockRejectedValue(new TypeError("fetch failed"));

    const response = await GET(request("http://console/upstream/api/v1/health"), context("api", "v1", "health"));

    expect(response.status).toBe(502);
    // Deliberately bare. The client reads a 5xx without an error envelope as
    // "the API could not be reached", which is what makes the console show
    // the start-the-stack notice instead of a gateway error nobody can act on.
    expect(await response.text()).toBe("");
  });
});
