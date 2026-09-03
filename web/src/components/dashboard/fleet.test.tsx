import { screen, within } from "@testing-library/react";

import { fleetSchema } from "@/lib/api/schema";
import { toFleet } from "@/lib/telemetry/from-live";
import type { FleetView } from "@/lib/telemetry/types";
import { fleetFixture, gatewayStatusFixture, plain, renderWithLocale } from "@/test/utils";

import { FleetPanel } from "./fleet";

const fleet = () => toFleet(fleetSchema.parse(fleetFixture()));

function fleetOf(...devices: Record<string, unknown>[]): FleetView {
  return toFleet(fleetSchema.parse({ devices: devices.map((d) => gatewayStatusFixture(d)) }));
}

/** The table, scoped so a device name does not match the summary badge. */
const table = () => within(screen.getByRole("table"));

describe("FleetPanel", () => {
  it("heads every column with something distinct", () => {
    const { dict } = renderWithLocale(<FleetPanel fleet={fleet()} />);

    const headers = screen.getAllByRole("columnheader").map((cell) => cell.textContent);
    expect(headers).toEqual([
      dict.fleet.device,
      dict.fleet.target,
      dict.fleet.state,
      dict.fleet.clock,
      dict.fleet.skew,
      dict.fleet.availability,
      dict.fleet.samples,
    ]);
    // Two columns headed the same thing is worse than an unlabelled one: the
    // reader stops trusting the header row.
    expect(new Set(headers).size).toBe(headers.length);
  });

  it("lists every device with its address", () => {
    renderWithLocale(<FleetPanel fleet={fleet()} />);

    expect(table().getByText("TTR20 alpha")).toBeInTheDocument();
    expect(table().getByText("TTR20 beta")).toBeInTheDocument();
    expect(table().getByText("127.0.0.1:9001")).toBeInTheDocument();
  });

  it("names both states of each device", () => {
    const { dict } = renderWithLocale(<FleetPanel fleet={fleet()} />);

    expect(table().getByText(dict.states.degraded)).toBeInTheDocument();
    expect(table().getByText(dict.states.online)).toBeInTheDocument();
    expect(table().getByText(dict.states.warn)).toBeInTheDocument();
    expect(table().getByText(dict.states.ok)).toBeInTheDocument();
  });

  it("keeps the sign on the skew", () => {
    renderWithLocale(<FleetPanel fleet={fleet()} />);

    // A meter running fast and one running slow are different problems.
    expect(plain(table().getByText(/-2,50/).textContent ?? "")).toContain("-2,50");
    expect(table().getByText(/\+12/)).toBeInTheDocument();
  });

  it("says nothing rather than zero for a device with no samples", () => {
    const { dict } = renderWithLocale(
      <FleetPanel fleet={fleetOf({ clock: { state: "unknown", samples: 0, skewMs: 0 } })} />,
    );

    // "0 ms" would claim the clock had been measured and found correct.
    expect(table().getByText(dict.common.none)).toBeInTheDocument();
  });

  it("calls out the worst clock in the header", () => {
    const { dict } = renderWithLocale(<FleetPanel fleet={fleet()} />);

    // On a bench with three devices sorting by eye is free; on forty it is not.
    const badge = screen.getByText(dict.fleet.worst).parentElement!;
    expect(within(badge).getByText(/alpha/)).toBeInTheDocument();
  });

  it("has no worst-clock badge when nothing has been measured", () => {
    const { dict } = renderWithLocale(
      <FleetPanel fleet={toFleet(fleetSchema.parse({ devices: [] }))} />,
    );

    expect(screen.queryByText(dict.fleet.worst)).not.toBeInTheDocument();
    expect(screen.getByText(dict.fleet.empty)).toBeInTheDocument();
  });

  it("explains a fleet of one", () => {
    const { dict } = renderWithLocale(<FleetPanel fleet={fleetOf({ deviceId: "solo" })} />);

    // A gateway with no inventory reports a fleet of one, and a single-row
    // table with no explanation reads as a fleet that lost its devices.
    expect(screen.getByText(dict.fleet.single)).toBeInTheDocument();
  });

  it("says nothing about inventories when there is more than one device", () => {
    const { dict } = renderWithLocale(<FleetPanel fleet={fleet()} />);

    expect(screen.queryByText(dict.fleet.single)).not.toBeInTheDocument();
  });

  it("marks which devices are being polled", () => {
    const { container } = renderWithLocale(
      <FleetPanel fleet={fleetOf({ deviceId: "up", state: "running" }, { deviceId: "down", state: "stopped" })} />,
    );

    expect(container.querySelectorAll(".pulse-dot")).toHaveLength(1);
  });

  it("translates", () => {
    const { dict } = renderWithLocale(<FleetPanel fleet={fleet()} />, { locale: "en" });

    expect(screen.getByRole("heading", { name: dict.fleet.title })).toBeInTheDocument();
    expect(table().getByText(dict.states.degraded)).toBeInTheDocument();
  });
});
