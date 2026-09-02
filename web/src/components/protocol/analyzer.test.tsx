import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { buildReadTimeRequest, encodeFrame, formatHex } from "@/lib/ft12";
import { renderWithLocale } from "@/test/utils";

import { FrameAnalyzer } from "./analyzer";

const REQUEST_SUM = encodeFrame(0x00, 0x01, buildReadTimeRequest(), "sum");

const inputOf = (dict: ReturnType<typeof renderWithLocale>["dict"]) =>
  screen.getByLabelText(dict.protocol.input);

describe("FrameAnalyzer", () => {
  it("starts on the canonical read-time request", () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    expect(inputOf(dict)).toHaveValue(formatHex(REQUEST_SUM));
    expect(screen.getByText(dict.protocol.valid)).toBeInTheDocument();
  });

  it("reports the frame and payload lengths", () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    expect(screen.getByText(dict.protocol.frameLength)).toBeInTheDocument();
    expect(screen.getByText("8")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("decodes the command inside the frame", () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    expect(screen.getByText("read-time")).toBeInTheDocument();
    // "запрос" also labels the builder's direction control, so the assertion is
    // that the decode badge exists alongside it rather than instead of it.
    expect(screen.getAllByText(dict.common.request).length).toBeGreaterThanOrEqual(2);
  });

  it("flags a frame the moment a byte is broken", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);
    const input = inputOf(dict);

    await userEvent.clear(input);
    await userEvent.type(input, "68 03 68 00 01 01 FF 16");

    expect(screen.getByText(dict.protocol.errors.invalidChecksum)).toBeInTheDocument();
  });

  it("prompts when the input is emptied", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.click(screen.getByRole("button", { name: dict.common.clear }));

    expect(screen.getByText(dict.protocol.empty)).toBeInTheDocument();
  });

  it("re-reads the same bytes when the checksum mode changes", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.click(screen.getByRole("radio", { name: "crc16" }));

    expect(screen.getByText(dict.protocol.errors.tooShort)).toBeInTheDocument();
  });

  it("loads a sample frame", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.click(screen.getByRole("button", { name: "read-identity · response" }));

    expect(screen.getByText("read-identity")).toBeInTheDocument();
    expect(screen.getByText("TTR20")).toBeInTheDocument();
    expect(screen.getAllByText(dict.common.response).length).toBeGreaterThanOrEqual(2);
  });

  it("loads the deliberately broken sample", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.click(screen.getByRole("button", { name: "bad checksum" }));

    expect(screen.getByText(dict.protocol.errors.invalidChecksum)).toBeInTheDocument();
  });

  it("warns about tokens that are not hex", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);
    const input = inputOf(dict);

    await userEvent.clear(input);
    await userEvent.type(input, "68 zz 68");

    expect(screen.getByText("zz")).toBeInTheDocument();
  });
});

describe("FrameAnalyzer builder", () => {
  it("writes the built frame into the decode box", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.selectOptions(screen.getByLabelText(dict.protocol.command), "2");
    await userEvent.click(screen.getByRole("button", { name: dict.protocol.insert }));

    expect(screen.getByText("read-identity")).toBeInTheDocument();
  });

  it("sets the response bit when the direction is a response", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.click(screen.getByRole("radio", { name: dict.common.response }));
    await userEvent.click(screen.getByRole("button", { name: dict.protocol.insert }));

    expect(screen.getAllByText(dict.common.response).length).toBeGreaterThan(0);
    expect(screen.getByText("80")).toBeInTheDocument();
  });

  it("rebuilds when the address changes", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.selectOptions(screen.getByLabelText(dict.protocol.address), "255");
    await userEvent.click(screen.getByRole("button", { name: dict.protocol.insert }));

    expect(inputOf(dict)).toHaveValue(
      formatHex(encodeFrame(0x00, 0xff, buildReadTimeRequest(), "sum")),
    );
  });

  it("rebuilds when the control byte changes", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.selectOptions(screen.getByLabelText(dict.protocol.control), "3");
    await userEvent.click(screen.getByRole("button", { name: dict.protocol.insert }));

    expect(inputOf(dict)).toHaveValue(
      formatHex(encodeFrame(0x03, 0x01, buildReadTimeRequest(), "sum")),
    );
  });
});

describe("FrameAnalyzer stream scan", () => {
  it("appears only once the input holds more than one frame", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    expect(screen.queryByText(dict.protocol.streamTitle)).not.toBeInTheDocument();

    const input = inputOf(dict);
    await userEvent.clear(input);
    await userEvent.paste(`${formatHex(REQUEST_SUM)} ${formatHex(REQUEST_SUM)}`);

    expect(screen.getByText(dict.protocol.streamTitle)).toBeInTheDocument();
    expect(screen.getByText("#1")).toBeInTheDocument();
    expect(screen.getByText("#2")).toBeInTheDocument();
  });

  it("reports the bytes left over after the last frame", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);
    const input = inputOf(dict);

    await userEvent.clear(input);
    await userEvent.paste(`${formatHex(REQUEST_SUM)} ${formatHex(REQUEST_SUM)} 68 03 68`);

    expect(screen.getByText(new RegExp(dict.protocol.streamRest))).toBeInTheDocument();
  });
});
