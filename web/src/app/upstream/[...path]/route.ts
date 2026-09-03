import { NextResponse, type NextRequest } from "next/server";

/**
 * The proxy to the Go HTTP API.
 *
 * The browser only ever talks to this app's own origin: it keeps the API off
 * the public network in a deployment, removes the CORS negotiation entirely,
 * and means the page never needs to know where the API lives.
 *
 * ## Why a route handler and not a rewrite
 *
 * This was `next.config.ts`'s `rewrites()`, which is simpler and wrong for a
 * container. Next resolves a rewrite destination at *build* time and writes it
 * into `routes-manifest.json`, so an image built with the default would point
 * at `http://127.0.0.1:8080` forever -- inside a container, that is the
 * container itself. `FT12_API_URL` would be read during `docker build` and
 * ignored at `docker run`, which is the opposite of what an image is for.
 *
 * A handler reads the variable per request, so one image runs anywhere.
 */

/** Where the API is, read fresh so a container can be told at start-up. */
function upstream(): string {
  return (process.env.FT12_API_URL ?? "http://127.0.0.1:8080").replace(/\/+$/, "");
}

/**
 * How long the console will wait for the API.
 *
 * Longer than any request the API makes of its own upstreams (3s by default),
 * short enough that a hung backend shows as unreachable rather than as a page
 * that never settles.
 */
const TIMEOUT_MS = 10_000;

/**
 * Headers that describe the hop rather than the message, and must not be
 * copied onto the next one. `host` in particular would send the API a name it
 * does not answer to.
 */
const HOP_BY_HOP = new Set([
  "connection",
  "content-encoding",
  "content-length",
  "host",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
]);

function forwardHeaders(source: Headers): Headers {
  const out = new Headers();
  source.forEach((value, key) => {
    if (!HOP_BY_HOP.has(key.toLowerCase())) out.set(key, value);
  });
  return out;
}

async function proxy(request: NextRequest, path: string[]): Promise<Response> {
  const target = `${upstream()}/${path.join("/")}${request.nextUrl.search}`;

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);

  try {
    const response = await fetch(target, {
      method: request.method,
      headers: forwardHeaders(request.headers),
      // GET and HEAD carry no body, and passing one makes fetch throw.
      body: request.method === "GET" || request.method === "HEAD" ? undefined : await request.text(),
      // The console shows what the gateway is doing now; a cached answer is
      // worse than none, because it looks current.
      cache: "no-store",
      signal: controller.signal,
      redirect: "manual",
    });

    return new NextResponse(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: forwardHeaders(response.headers),
    });
  } catch {
    // Deliberately bare: no JSON error envelope.
    //
    // The API's own failures always carry `{error:{code,message}}`, and the
    // client reads the absence of that envelope as "the API could not be
    // reached at all" -- which is exactly what has happened here, and what
    // makes the console show the start-the-stack notice rather than a gateway
    // error nobody can act on.
    return new NextResponse(null, { status: 502, statusText: "Bad Gateway" });
  } finally {
    clearTimeout(timer);
  }
}

type Context = { params: Promise<{ path: string[] }> };

export async function GET(request: NextRequest, context: Context): Promise<Response> {
  return proxy(request, (await context.params).path);
}

export async function POST(request: NextRequest, context: Context): Promise<Response> {
  return proxy(request, (await context.params).path);
}

export async function PUT(request: NextRequest, context: Context): Promise<Response> {
  return proxy(request, (await context.params).path);
}

/**
 * Nothing is prerendered: every answer describes a device as it is right now,
 * and a cached one would describe it as it was when the image was built.
 */
export const dynamic = "force-dynamic";
