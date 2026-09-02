/**
 * Jest, SWC-based — no Babel and no Vite, matching the workspace this stack
 * comes from.
 *
 * `runtime: "automatic"` means components need no `import React`, the same as
 * the app's `jsx: "preserve"` handed to Next.
 */

/** @type {import("jest").Config} */
const config = {
  testEnvironment: "jsdom",
  setupFilesAfterEnv: ["<rootDir>/jest.setup.ts"],
  transform: {
    "^.+\.(t|j)sx?$": [
      "@swc/jest",
      {
        jsc: {
          parser: { syntax: "typescript", tsx: true },
          target: "es2022",
          transform: { react: { runtime: "automatic", development: false } },
        },
      },
    ],
  },
  moduleNameMapper: {
    "^@/(.*)$": "<rootDir>/src/$1",
    // Stylesheets carry no behaviour worth asserting in a unit test.
    "\.(css|less|scss|sass)$": "<rootDir>/test/style-mock.cjs",
    // lucide-react's package entry is ESM, which Jest cannot require. It ships
    // a CommonJS build alongside it, so point at that rather than transforming
    // the whole of node_modules for one dependency.
    "^lucide-react$": "<rootDir>/node_modules/lucide-react/dist/cjs/lucide-react.js",
  },
  testMatch: ["<rootDir>/src/**/*.test.ts", "<rootDir>/src/**/*.test.tsx"],
  collectCoverageFrom: [
    "src/**/*.{ts,tsx}",
    "!src/**/*.test.{ts,tsx}",
    // Route files are one-line re-exports of the component they render; the
    // component is what carries behaviour and is tested directly.
    "!src/app/**/page.tsx",
    "!src/app/**/layout.tsx",
    "!src/app/not-found.tsx",
  ],
  clearMocks: true,
  restoreMocks: true,
};

export default config;
