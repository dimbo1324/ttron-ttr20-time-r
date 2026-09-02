import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { crc16, formatHex, sum8 } from "@/lib/ft12";
import { renderWithLocale } from "@/test/utils";

import { ProtocolReference } from "./reference";

const inputOf = (dict: ReturnType<typeof renderWithLocale>["dict"]) =>
  screen.getByLabelText(dict.reference.calculatorInput);

describe("ProtocolReference", () => {
  it("states the frame format", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    expect(screen.getByText(/0x68 \| LEN \| 0x68/)).toBeInTheDocument();
    expect(screen.getByText(dict.reference.frameFormatBody)).toBeInTheDocument();
  });

  it("describes both checksum modes", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    expect(screen.getByText(dict.reference.checksumSum)).toBeInTheDocument();
    expect(screen.getByText(dict.reference.checksumCrc)).toBeInTheDocument();
  });

  it("carries the time-zone warning", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    expect(screen.getByText(dict.reference.zoneTitle)).toBeInTheDocument();
    expect(screen.getByText(dict.reference.zoneBody)).toBeInTheDocument();
  });

  it("lists both commands with their wire formats", () => {
    renderWithLocale(<ProtocolReference />);

    expect(screen.getByText("read-time")).toBeInTheDocument();
    expect(screen.getByText("read-identity")).toBeInTheDocument();
    expect(screen.getByText('DATA = 0x01 + "YYYY-MM-DD HH:MM:SS"')).toBeInTheDocument();
    expect(screen.getByText('DATA = 0x02 + "MODEL|SERIAL|FIRMWARE"')).toBeInTheDocument();
  });

  it("lists the decode error vocabulary", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    expect(screen.getByText("invalidChecksum")).toBeInTheDocument();
    expect(screen.getByText(dict.protocol.errors.invalidChecksum)).toBeInTheDocument();
  });
});

describe("ProtocolReference calculator", () => {
  it("computes both checksums of the default payload", () => {
    renderWithLocale(<ProtocolReference />);

    // The documented vector: sum8([0x00,0x01,0x01]) === 0x02.
    expect(screen.getByText("02")).toBeInTheDocument();
    expect(screen.getByText(formatHex([crc16([0, 1, 1]) & 0xff, (crc16([0, 1, 1]) >> 8) & 0xff])))
      .toBeInTheDocument();
  });

  it("recomputes as the payload is edited", async () => {
    const { dict } = renderWithLocale(<ProtocolReference />);
    const input = inputOf(dict);

    await userEvent.clear(input);
    await userEvent.paste("FF 02");

    const expectedSum = sum8([0xff, 0x02]).toString(16).toUpperCase().padStart(2, "0");
    expect(screen.getByText(expectedSum)).toBeInTheDocument();
  });

  it("handles an empty payload without failing", async () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    await userEvent.clear(inputOf(dict));

    expect(screen.getByText("00")).toBeInTheDocument();
  });

  it("redraws the sample frame when the mode changes", async () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    await userEvent.click(screen.getByRole("radio", { name: "crc16" }));

    expect(screen.getAllByText(dict.protocol.fields.checksum).length).toBeGreaterThan(0);
  });
});
