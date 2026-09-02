import type { NextConfig } from "next";

/**
 * The console talks to the Go HTTP API (`cmd/ft12-api`, default `:8080`).
 *
 * Requests are proxied through this app rather than sent to the API origin
 * from the browser: it keeps the API off the public network in a deployment,
 * removes the CORS negotiation entirely, and means the browser only ever
 * needs one origin. `FT12_API_URL` points at the API from the server side.
 */
const apiUrl = process.env.FT12_API_URL ?? "http://127.0.0.1:8080";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      { source: "/upstream/:path*", destination: `${apiUrl}/:path*` },
    ];
  },
};

export default nextConfig;
