import { render, screen } from "@testing-library/react";

import { useMounted } from "./use-mounted";

function Probe() {
  return <span data-testid="probe">{useMounted() ? "mounted" : "pending"}</span>;
}

describe("useMounted", () => {
  it("is true once the effect has run", () => {
    render(<Probe />);

    expect(screen.getByTestId("probe")).toHaveTextContent("mounted");
  });

  it("starts false so the server and first client render agree", () => {
    // React runs effects after the initial commit; reading the value during
    // that first render is what the hook exists to guard.
    let firstRender: boolean | undefined;

    function Capture() {
      const mounted = useMounted();
      firstRender ??= mounted;
      return null;
    }

    render(<Capture />);
    expect(firstRender).toBe(false);
  });
});
