import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { api } from "@/lib/api/client";
import { renderWithLocale, resetLiveStore, useSource } from "@/test/utils";
import { useSourceStore } from "@/stores/source-store";
import { useLiveStore } from "@/stores/live-store";

import { SourceNotice } from "./source-notice";
import { SourceSwitch } from "./source-switch";

/**
 * The two controls that tell an operator which stack they are looking at, and
 * what to do when it is not there.
 */

jest.mock("@/lib/api/client", () => {
  const actual = jest.requireActual<typeof import("@/lib/api/client")>("@/lib/api/client");
  return {
    ...actual,
    api: { faultMode: jest.fn(), gatewayStatus: jest.fn() },
  };
});

void api;

beforeEach(() => {
  window.localStorage.clear();
  resetLiveStore();
  useSourceStore.setState({ source: "bench", hydrated: true });
});

describe("SourceSwitch", () => {
  it("names both sources, not their explanations", () => {
    const { dict } = renderWithLocale(<SourceSwitch link="ready" />);

    // The hint is a tooltip. Without an explicit label the title attribute
    // becomes the accessible name, and a screen reader reads a paragraph
    // where it should say "Bench".
    expect(screen.getByRole("radio", { name: dict.source.bench })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: dict.source.live })).toBeInTheDocument();
  });

  it("marks the active source", () => {
    const { dict } = renderWithLocale(<SourceSwitch link="ready" />);

    expect(screen.getByRole("radio", { name: dict.source.bench })).toHaveAttribute(
      "aria-checked",
      "true",
    );
    expect(screen.getByRole("radio", { name: dict.source.live })).toHaveAttribute(
      "aria-checked",
      "false",
    );
  });

  it("switches source on click", async () => {
    const { dict } = renderWithLocale(<SourceSwitch link="ready" />);

    await userEvent.click(screen.getByRole("radio", { name: dict.source.live }));

    expect(useSourceStore.getState().source).toBe("live");
  });

  it("says nothing about a link while the bench is selected", () => {
    const { dict } = renderWithLocale(<SourceSwitch link="unreachable" />);

    // The bench cannot fail to be reachable; a link badge beside it would be
    // meaningless and, worse, alarming.
    expect(screen.queryByText(dict.source.linkUnreachable)).not.toBeInTheDocument();
  });

  it.each([
    ["ready", "linkReady"],
    ["connecting", "linkConnecting"],
    ["unreachable", "linkUnreachable"],
  ] as const)("shows the %s link on the live source", (link, key) => {
    useSourceStore.setState({ source: "live" });
    const { dict } = renderWithLocale(<SourceSwitch link={link} />);

    expect(screen.getByText(dict.source[key])).toBeInTheDocument();
  });

  it("translates", () => {
    const { dict } = renderWithLocale(<SourceSwitch link="ready" />, { locale: "en" });

    expect(screen.getByRole("radio", { name: dict.source.live })).toBeInTheDocument();
  });
});

describe("SourceNotice", () => {
  it("says nothing on the bench", () => {
    const { container } = renderWithLocale(<SourceNotice />);

    expect(container).toBeEmptyDOMElement();
  });

  it("says nothing on a healthy live source", () => {
    useSource("live");
    useLiveStore.setState({ link: "ready", error: null });

    const { container } = renderWithLocale(<SourceNotice />);

    expect(container).toBeEmptyDOMElement();
  });

  it("names the missing processes and how to start them", () => {
    useSource("live");
    useLiveStore.setState({ link: "unreachable", error: "Failed to fetch" });

    const { dict } = renderWithLocale(<SourceNotice />);

    expect(screen.getByText(dict.source.unreachableTitle)).toBeInTheDocument();
    expect(screen.getByText(dict.source.unreachableBody)).toBeInTheDocument();
    // Three separate commands, not chained: the operator may be in PowerShell,
    // where the POSIX way of backgrounding them is a syntax error.
    const command = screen.getByText(/ft12-emulator/);
    expect(command.textContent).toContain("go run ./cmd/ft12-gateway");
    expect(command.textContent).toContain("go run ./cmd/ft12-api");
    expect(command.textContent).not.toContain("&");
  });

  it("shows the underlying failure alongside the instructions", () => {
    useSource("live");
    useLiveStore.setState({ link: "unreachable", error: "Failed to fetch" });

    renderWithLocale(<SourceNotice />);

    expect(screen.getByText("Failed to fetch")).toBeInTheDocument();
  });

  it("offers the command to the clipboard", async () => {
    useSource("live");
    useLiveStore.setState({ link: "unreachable", error: null });
    const writeText = jest.fn(async () => undefined);
    Object.assign(navigator.clipboard, { writeText });

    const { dict } = renderWithLocale(<SourceNotice />);
    await userEvent.click(screen.getByRole("button", { name: dict.common.copy }));

    expect(writeText).toHaveBeenCalledWith(expect.stringContaining("ft12-api"));
  });

  it("gives a reachable gateway's error a quieter band", () => {
    useSource("live");
    useLiveStore.setState({ link: "ready", error: "dial tcp: connection refused" });

    const { dict } = renderWithLocale(<SourceNotice />);

    // The readings below are real and still worth looking at, so this is a
    // band rather than a takeover.
    expect(screen.getByText(dict.source.upstreamError)).toBeInTheDocument();
    expect(screen.getByText("dial tcp: connection refused")).toBeInTheDocument();
    expect(screen.queryByText(dict.source.unreachableTitle)).not.toBeInTheDocument();
  });

  it("translates", () => {
    useSource("live");
    useLiveStore.setState({ link: "unreachable", error: null });

    const { dict } = renderWithLocale(<SourceNotice />, { locale: "en" });

    expect(screen.getByText(dict.source.unreachableTitle)).toBeInTheDocument();
  });
});
