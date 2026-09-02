import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildReadIdentityResponse,
  buildReadTimeRequest,
  buildReadTimeResponse,
  decodeFrame,
  encodeFrame,
  RESPONSE_BIT,
} from "@/lib/ft12";
import { renderWithLocale } from "@/test/utils";

import { ByteGrid, CommandDecode, FrameInspector, FrameLayoutBar, FrameVerdict } from "./frame-view";

const REQUEST = encodeFrame(0x00, 0x01, buildReadTimeRequest(), "sum");
const RESPONSE = encodeFrame(
  RESPONSE_BIT,
  0x01,
  buildReadTimeResponse(new Date(2026, 5, 2, 12, 34, 56)),
  "sum",
);

const corrupted = () => {
  const raw = [...REQUEST];
  raw[raw.length - 2] = raw[raw.length - 2]! ^ 0xff;
  return raw;
};

describe("FrameLayoutBar", () => {
  it("names every field in the legend", () => {
    const { dict } = renderWithLocale(<FrameLayoutBar decoded={decodeFrame(REQUEST, "sum")} />);

    for (const label of [
      dict.protocol.fields.start,
      dict.protocol.fields.length,
      dict.protocol.fields.control,
      dict.protocol.fields.address,
      dict.protocol.fields.data,
      dict.protocol.fields.checksum,
      dict.protocol.fields.end,
    ]) {
      expect(screen.getAllByText(label).length).toBeGreaterThan(0);
    }
  });

  it("renders nothing for a frame with no fields", () => {
    const { container } = renderWithLocale(
      <FrameLayoutBar decoded={decodeFrame([0x68], "sum")} />,
    );

    expect(container).toBeEmptyDOMElement();
  });
});

describe("ByteGrid", () => {
  it("renders one cell per byte with its offset", () => {
    renderWithLocale(<ByteGrid decoded={decodeFrame(REQUEST, "sum")} />);

    expect(screen.getAllByRole("button")).toHaveLength(REQUEST.length);
    expect(screen.getAllByText("68").length).toBeGreaterThan(0);
    expect(screen.getByText("16")).toBeInTheDocument();
  });

  it("reveals the binary and ASCII of a byte on hover", async () => {
    const { dict } = renderWithLocale(<ByteGrid decoded={decodeFrame(REQUEST, "sum")} />);

    expect(screen.getByText(dict.protocol.byteMap)).toBeInTheDocument();

    await userEvent.hover(screen.getAllByRole("button")[0]!);

    expect(screen.getByText("0x68")).toBeInTheDocument();
    expect(screen.getByText("01101000")).toBeInTheDocument();
    expect(screen.getByText(dict.protocol.fields.start)).toBeInTheDocument();
  });

  it("returns to the idle strip when the pointer leaves", async () => {
    const { dict } = renderWithLocale(<ByteGrid decoded={decodeFrame(REQUEST, "sum")} />);
    const cell = screen.getAllByRole("button")[0]!;

    await userEvent.hover(cell);
    await userEvent.unhover(cell);

    expect(screen.getByText(dict.protocol.byteMap)).toBeInTheDocument();
  });

  it("renders nothing when there are no bytes to map", () => {
    const { container } = renderWithLocale(<ByteGrid decoded={decodeFrame([], "sum")} />);

    expect(container).toBeEmptyDOMElement();
  });
});

describe("FrameVerdict", () => {
  it("confirms a valid frame", () => {
    const { dict } = renderWithLocale(<FrameVerdict decoded={decodeFrame(REQUEST, "sum")} />);

    expect(screen.getByText(dict.protocol.valid)).toBeInTheDocument();
  });

  it("explains a checksum mismatch and shows both values", () => {
    const { dict } = renderWithLocale(<FrameVerdict decoded={decodeFrame(corrupted(), "sum")} />);

    expect(screen.getByText(dict.protocol.errors.invalidChecksum)).toBeInTheDocument();
    expect(screen.getByText(/FD/)).toBeInTheDocument();
  });

  it("explains a bad start byte", () => {
    const raw = [...REQUEST];
    raw[0] = 0x00;
    const { dict } = renderWithLocale(<FrameVerdict decoded={decodeFrame(raw, "sum")} />);

    expect(screen.getByText(dict.protocol.errors.invalidStart)).toBeInTheDocument();
  });
});

describe("CommandDecode", () => {
  it("reads a request", () => {
    const { dict } = renderWithLocale(<CommandDecode decoded={decodeFrame(REQUEST, "sum")} />);

    expect(screen.getByText(dict.common.request)).toBeInTheDocument();
    expect(screen.getByText("read-time")).toBeInTheDocument();
    expect(screen.getByText("0x01")).toBeInTheDocument();
  });

  it("reads a response and renders its timestamp", () => {
    const { dict } = renderWithLocale(<CommandDecode decoded={decodeFrame(RESPONSE, "sum")} />);

    expect(screen.getByText(dict.common.response)).toBeInTheDocument();
    expect(screen.getByText("2026-06-02 12:34:56")).toBeInTheDocument();
  });

  it("splits an identity response into labelled fields", () => {
    const frame = encodeFrame(
      RESPONSE_BIT,
      0x01,
      buildReadIdentityResponse("TTR20", "SN-42", "1.2.3"),
      "sum",
    );
    renderWithLocale(<CommandDecode decoded={decodeFrame(frame, "sum")} />);

    expect(screen.getByText("model")).toBeInTheDocument();
    expect(screen.getByText("TTR20")).toBeInTheDocument();
    expect(screen.getByText("SN-42")).toBeInTheDocument();
    expect(screen.getByText("1.2.3")).toBeInTheDocument();
  });

  it("reports an unusable payload", () => {
    const frame = encodeFrame(RESPONSE_BIT, 0x01, [0x01, 0x30, 0x30], "sum");
    const { dict } = renderWithLocale(<CommandDecode decoded={decodeFrame(frame, "sum")} />);

    expect(screen.getByText(dict.protocol.errors.invalidLengthPayload)).toBeInTheDocument();
  });

  it("renders nothing without a control byte", () => {
    const { container } = renderWithLocale(<CommandDecode decoded={decodeFrame([0x68], "sum")} />);

    expect(container).toBeEmptyDOMElement();
  });
});

describe("FrameInspector", () => {
  it("prompts for input when there is no frame", () => {
    const { dict } = renderWithLocale(<FrameInspector bytes={[]} mode="sum" />);

    expect(screen.getByText(dict.protocol.empty)).toBeInTheDocument();
  });

  it("stacks the verdict, the map and the decode", () => {
    const { dict } = renderWithLocale(<FrameInspector bytes={REQUEST} mode="sum" />);

    expect(screen.getByText(dict.protocol.valid)).toBeInTheDocument();
    expect(screen.getByText("read-time")).toBeInTheDocument();
    expect(screen.getByText(dict.protocol.computed)).toBeInTheDocument();
    expect(screen.getAllByText(dict.protocol.fields.data).length).toBeGreaterThan(0);
  });

  it("drops the layout bar in compact mode", () => {
    const { dict } = renderWithLocale(<FrameInspector bytes={REQUEST} mode="sum" compact />);

    // The legend is the layout bar's only text, so its absence is the check.
    expect(screen.queryByText(dict.protocol.fields.startRepeat)).not.toBeInTheDocument();
    expect(screen.getByText(dict.protocol.valid)).toBeInTheDocument();
  });

  it("reads the same bytes differently in crc16 mode", () => {
    const { dict } = renderWithLocale(<FrameInspector bytes={REQUEST} mode="crc16" />);

    expect(screen.getByText(dict.protocol.errors.tooShort)).toBeInTheDocument();
  });
});

describe("ByteGrid keyboard access", () => {
  it("inspects a byte on focus and clears it on blur", async () => {
    const { dict } = renderWithLocale(<ByteGrid decoded={decodeFrame(REQUEST, "sum")} />);
    const cells = screen.getAllByRole("button");

    await userEvent.tab();
    expect(cells[0]).toHaveFocus();
    expect(screen.getByText("0x68")).toBeInTheDocument();

    await userEvent.tab();
    expect(screen.getByText("0x03")).toBeInTheDocument();
    expect(screen.queryByText(dict.protocol.byteMap)).not.toBeInTheDocument();
  });
});

describe("FrameVerdict marks the offending byte", () => {
  it("rings the checksum cell of a corrupted frame", () => {
    const { container } = renderWithLocale(
      <ByteGrid decoded={decodeFrame(corrupted(), "sum")} />,
    );

    expect(container.querySelectorAll('[class*="ring-destructive"]')).toHaveLength(1);
  });
});
