/**
 * Captures the screenshots and animation frames used by the documentation.
 *
 * Docs that show an interface go stale the way docs that describe one do, and
 * they go stale invisibly: nobody diffs a PNG. So the pictures are produced by
 * a script against a running stack rather than by a person with a snipping
 * tool, and re-running them after a change is a two-minute job instead of an
 * afternoon.
 *
 * Stills are written straight to docs/media/. Animations are written as
 * numbered PNG frames into a scratch directory and folded into a GIF by
 * `tools/docmedia/gif` -- see tools/docmedia/README.md.
 *
 *   node tools/docmedia/capture.mjs --base http://127.0.0.1:3000
 *   node tools/docmedia/capture.mjs --only overview,monitor
 *
 * Chrome is used where it is installed rather than downloaded: this script is
 * run by hand, occasionally, and a browser download is a poor trade for that.
 */

import { existsSync, mkdirSync, rmSync } from "node:fs";
import path from "node:path";
import process from "node:process";

import { chromium } from "playwright-core";

// -------------------------------------------------------------------- config

const CHROME_CANDIDATES = [
  "C:/Program Files/Google/Chrome/Application/chrome.exe",
  "C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe",
  "/usr/bin/google-chrome",
  "/usr/bin/chromium",
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
];

const args = parseArgs(process.argv.slice(2));
const BASE = args.base ?? "http://127.0.0.1:3000";
const OUT = args.out ?? "docs/media";
const FRAMES = args.frames ?? path.join(OUT, ".frames");
const ONLY = args.only ? new Set(args.only.split(",")) : null;

/** Stills. Two-times scale: the text in this interface is small by design. */
const VIEWPORT = { width: 1360, height: 860 };
const SCALE = 2;

/** Animations. Smaller and unscaled -- every pixel here is paid for twice. */
const CLIP = { width: 1120, height: 660 };

/**
 * How long the bench runs before anything is photographed.
 *
 * The charts are windows over recent history, and a console photographed the
 * moment it started is a page of dashes. This is the one place the capture is
 * deliberately slow.
 */
const WARMUP_MS = 45_000;

// --------------------------------------------------------------------- shots

/** Pages, in the order the navigation lists them. */
const PAGES = {
  overview: "Overview",
  monitor: "Monitor",
  frames: "Frames",
  emulator: "Emulator",
  gateway: "Gateway",
  reference: "Reference",
};

const BENCH_STILLS = [
  { name: "overview", page: "overview", note: "the dashboard" },
  { name: "monitor", page: "monitor", note: "the exchange monitor" },
  { name: "frames", page: "frames", note: "the frame analyzer" },
  { name: "emulator", page: "emulator", note: "fault modes and scenarios" },
  { name: "gateway", page: "gateway", note: "schedule, retries, clock, health" },
  { name: "reference", page: "reference", note: "the protocol reference" },
];

const LIVE_STILLS = [
  { name: "live-overview", page: "overview", note: "the same dashboard, reading a Go stack" },
  { name: "live-gateway", page: "gateway", note: "reconfiguring a gateway that is polling" },
  {
    name: "live-fleet",
    page: "overview",
    panel: "Device fleet",
    note: "every device the gateway polls",
  },
];

const ANIMATIONS = [
  {
    name: "monitor",
    note: "frames arriving, counters moving",
    async run(page, shot) {
      await goto(page, "/en");
      await start(page);
      await fastSchedule(page);
      await nav(page, "monitor");
      await page.waitForTimeout(12_000);
      await frames(page, shot, 40, 280);
    },
  },
  {
    name: "fault-injection",
    note: "a fault switched on, and the line degrading under it",
    // Wider than the others: the scenario buttons and the fault sliders they
    // move are in two different columns, and below this width the layout
    // stacks them -- so clicking one would scroll the other off screen and
    // the animation would show a page jump instead of a cause and an effect.
    clip: { width: 1340, height: 750 },
    async run(page, shot) {
      await goto(page, "/en");
      await start(page);
      await fastSchedule(page);
      await nav(page, "emulator");
      await page.waitForTimeout(8000);
      await frames(page, shot, 8, 280);
      await page.getByRole("button", { name: "Noisy line" }).click();
      await frames(page, shot, 38, 280);
    },
  },
  {
    name: "live-settings",
    source: "live",
    note: "a running gateway re-planning its schedule from the console",
    async run(page, shot) {
      await goto(page, "/en/gateway");
      await page.waitForTimeout(8000);
      await frames(page, shot, 8, 300);
      await page.locator("select").first().selectOption({ index: 4 });
      await frames(page, shot, 40, 300);
    },
  },
];

// ---------------------------------------------------------------------- main

const executablePath = CHROME_CANDIDATES.find((candidate) => existsSync(candidate));
if (!executablePath) {
  console.error("no Chrome or Edge found; looked for:\n  " + CHROME_CANDIDATES.join("\n  "));
  process.exit(1);
}

mkdirSync(OUT, { recursive: true });
const browser = await chromium.launch({ executablePath, headless: true });

// One session per source rather than one per screenshot. The bench keeps its
// history in the page, so a fresh context per shot would photograph six
// consoles that had each just been switched on.
if (wanted(BENCH_STILLS)) {
  const page = await open(browser, "bench", VIEWPORT, SCALE);
  await goto(page, "/en");
  await start(page);
  await warmUp(page);
  for (const still of BENCH_STILLS) await capture(page, still);
  await page.context().close();
}

if (wanted(LIVE_STILLS)) {
  const page = await open(browser, "live", VIEWPORT, SCALE);
  await goto(page, "/en");
  await page.waitForTimeout(12_000);
  for (const still of LIVE_STILLS) await capture(page, still);
  await page.context().close();
}

// Photographed with the stack down, so it has to be asked for by name.
if (ONLY?.has("unreachable")) {
  const page = await open(browser, "live", VIEWPORT, SCALE);
  await goto(page, "/en");
  await page.waitForTimeout(8000);
  await capture(page, { name: "unreachable", note: "the live source with nothing answering" });
  await page.context().close();
}

for (const animation of ANIMATIONS) {
  if (ONLY && !ONLY.has(animation.name)) continue;
  const dir = path.join(FRAMES, animation.name);
  rmSync(dir, { recursive: true, force: true });
  mkdirSync(dir, { recursive: true });

  const clip = animation.clip ?? CLIP;
  const page = await open(browser, animation.source ?? "bench", clip, 1);
  let frame = 0;
  const shot = () =>
    page.screenshot({
      path: path.join(dir, String(frame++).padStart(4, "0") + ".png"),
      clip: { x: 0, y: 0, ...clip },
    });

  await animation.run(page, shot);
  console.log(`${dir}  ${frame} frames -- ${animation.note}`);
  await page.context().close();
}

await browser.close();

// ------------------------------------------------------------------- helpers

function wanted(stills) {
  return !ONLY || stills.some((still) => ONLY.has(still.name));
}

async function capture(page, still) {
  if (ONLY && !ONLY.has(still.name)) return;
  if (still.page) await nav(page, still.page);
  await page.waitForTimeout(2500);

  const file = path.join(OUT, `${still.name}.png`);
  // A shot of one panel rather than the whole window, for the places the docs
  // are talking about one thing and the rest of the page is a distraction.
  const target = still.panel
    ? page
        .locator("section")
        .filter({ has: page.getByRole("heading", { name: still.panel, exact: true }) })
        .first()
    : page;
  if (still.panel) {
    await target.scrollIntoViewIfNeeded();
    await page.waitForTimeout(800);
  }
  await target.screenshot({ path: file });
  console.log(`${file}  ${still.note}`);
}

async function open(browser, source, viewport, scale) {
  const context = await browser.newContext({
    viewport,
    deviceScaleFactor: scale,
    colorScheme: "dark",
    // The console renders timestamps in the browser's zone, and a different
    // one in every capture reads as a change to the page.
    timezoneId: "UTC",
    locale: "en-US",
  });
  // The console remembers the chosen source per browser, so it is set before
  // the first script on the page runs rather than by clicking afterwards --
  // clicking would put the switch itself into the first frame.
  await context.addInitScript((value) => {
    try {
      window.localStorage.setItem("ft12.source", value);
    } catch {
      /* a browser blocking site data still gets the default */
    }
  }, source);
  const page = await context.newPage();
  // The development overlay is not part of the product.
  await page.addStyleTag({ content: "nextjs-portal { display: none !important; }" }).catch(() => {});
  return page;
}

async function goto(page, route) {
  // Not `networkidle`: on the live source the console polls continuously, and
  // with nothing answering it retries -- there is no idle moment to wait for.
  // Every shot has an explicit settling time anyway.
  await page.goto(BASE + route, { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(1500);
  await page.addStyleTag({ content: "nextjs-portal { display: none !important; }" }).catch(() => {});
}

/** Follows the navigation rather than reloading, so page state survives. */
async function nav(page, name) {
  await page.getByRole("link", { name: PAGES[name], exact: true }).click();
  await page.waitForTimeout(600);
}

async function start(page) {
  const button = page.getByRole("button", { name: "Start", exact: true });
  if (await button.count()) await button.first().click();
}

/**
 * Runs the bench fast enough that a minute of history fills a chart.
 *
 * The shipped default reads on the fifth second of every minute, which is the
 * brief and a poor photograph: a one-minute window would hold a single point.
 */
async function fastSchedule(page) {
  await nav(page, "gateway");
  await page.getByRole("radio", { name: "By interval" }).click().catch(() => {});
  await page.locator("select").first().selectOption({ index: 0 }).catch(() => {});
  await nav(page, "overview");
}

async function warmUp(page) {
  await fastSchedule(page);
  await page.waitForTimeout(WARMUP_MS);
}

async function frames(page, shot, count, everyMs) {
  for (let i = 0; i < count; i++) {
    await shot();
    await page.waitForTimeout(everyMs);
  }
}

function parseArgs(argv) {
  const out = {};
  for (let i = 0; i < argv.length; i++) {
    if (!argv[i].startsWith("--")) continue;
    out[argv[i].slice(2)] = argv[i + 1]?.startsWith("--") ? true : argv[++i];
  }
  return out;
}
