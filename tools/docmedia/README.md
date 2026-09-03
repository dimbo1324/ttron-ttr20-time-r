# Documentation media

The screenshots and animations in [docs/tour.md](../../docs/tour.md) are
produced by a script rather than by a person with a snipping tool.

Prose that has gone stale is at least readable and can be argued with. A
screenshot that has gone stale looks authoritative and is silently wrong, and
nobody diffs a PNG. Making the pictures reproducible turns "the interface
changed" from an afternoon into two minutes.

## What it produces

| File | What it shows |
| --- | --- |
| `overview.png` | the dashboard, bench source |
| `monitor.png` | the exchange monitor |
| `frames.png` | the frame analyzer |
| `emulator.png` | fault modes, scenarios, identity |
| `gateway.png` | schedule, retries, clock thresholds, availability |
| `reference.png` | the protocol reference |
| `live-overview.png` | the same dashboard reading a running Go stack |
| `live-gateway.png` | reconfiguring a gateway that is already polling |
| `live-fleet.png` | one gateway, three devices, one of them dead |
| `unreachable.png` | the live source with nothing answering |
| `monitor.gif` | frames arriving, counters moving |
| `fault-injection.gif` | a fault switched on, and the line degrading under it |
| `live-settings.gif` | a running gateway re-planning its schedule |

## Running it

The stack has to be up, because the live screenshots are of a real gateway
reading a real emulator:

```sh
docker compose up -d --build
```

Then, once:

```sh
npm --prefix tools/docmedia install
```

and:

```sh
node tools/docmedia/capture.mjs
```

Stills land in `docs/media/`. Animations land as numbered PNG frames in
`docs/media/.frames/<name>/`, which is gitignored, and are folded into GIFs by
a second pass:

```sh
go run ./tools/docmedia/gif -in docs/media/.frames/monitor -out docs/media/monitor.gif -delay 45
```

A single shot, by name:

```sh
node tools/docmedia/capture.mjs --only overview,live-fleet
```

## Two shots that need a different stack

Most of the set comes from a plain `docker compose up`. Two do not.

**`live-fleet.png`** wants more than one device. Point the gateway at the demo
inventory instead of a single `-target` -- see
[examples/devices.demo.json](../../examples/devices.demo.json), whose third
device is deliberately aimed at a port with nothing behind it -- and give it a
minute, because a device is not called offline until it has failed to answer
ten times running.

**`unreachable.png`** wants the console up and the API down:

```sh
docker compose stop ft12-api
node tools/docmedia/capture.mjs --only unreachable
docker compose start ft12-api
```

## Why GIF, and why not ffmpeg

A `.mp4` committed to a repository does not play inside a README on GitHub; it
downloads. An animated GIF plays wherever Markdown is rendered, which is the
only property that matters for documentation living in the tree.

`tools/docmedia/gif` encodes them with the standard library rather than
shelling out to ffmpeg, which is not a safe assumption about a machine. It
does the two things the standard library leaves to the caller: it picks 256
colours by median cut over all the frames at once, because the stock even
lattice spends its palette on colours a dark interface never uses and bands the
near-blacks it is mostly made of; and it writes only the rectangle that changed
between one frame and the next, because a console page is nearly all
background and re-encoding it forty times is most of the file.

## Determinism

The capture pins the browser's timezone to UTC and its locale to `en-US`, so a
re-run does not show up as a diff in every timestamp on the page. It also runs
the bench for forty-five seconds before photographing anything, at a one-second
interval rather than the shipped default: the charts are windows over recent
history, and the honest default -- a reading on the fifth second of every
minute -- would put one point in a one-minute window.

Chrome or Edge is used where it is already installed rather than downloaded.
This script runs by hand, occasionally; a browser download is a poor trade for
that.
