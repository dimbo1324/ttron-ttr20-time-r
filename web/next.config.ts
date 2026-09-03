import type { NextConfig } from "next";

/**
 * The console talks to the Go HTTP API (`cmd/ft12-api`, default `:8080`)
 * through a proxy on its own origin -- see `src/app/upstream/[...path]`.
 *
 * The proxy is a route handler rather than a `rewrites()` entry because Next
 * resolves a rewrite destination at build time. In a container that would
 * freeze the API address into the image, and `FT12_API_URL` would be read
 * during `docker build` and ignored at `docker run`.
 */
const nextConfig: NextConfig = {
  reactStrictMode: true,

  /**
   * A self-contained server directory, so the runtime image carries the app
   * and the handful of modules it actually imports rather than the whole
   * dependency tree. It is the difference between a couple of hundred
   * megabytes and a couple of dozen.
   */
  output: "standalone",
};

export default nextConfig;
