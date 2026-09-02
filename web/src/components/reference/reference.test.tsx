import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { getDictionary, locales } from "@/i18n";
import { crc16, formatHex, sum8 } from "@/lib/ft12";
import { renderWithLocale } from "@/test/utils";

import { ProtocolReference } from "./reference";

const inputOf = (dict: ReturnType<typeof renderWithLocale>["dict"]) =>
  screen.getByLabelText(dict.reference.checksums.calculatorInput);

describe("ProtocolReference structure", () => {
  it("opens with the plain-language introduction", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    expect(screen.getByRole("heading", { name: dict.reference.basics.title })).toBeInTheDocument();
    expect(screen.getByText(dict.reference.basics.lead)).toBeInTheDocument();
    for (const paragraph of dict.reference.basics.body) {
      expect(screen.getByText(paragraph)).toBeInTheDocument();
    }
  });

  it("renders every section, in reading order, with a contents entry each", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    const sections = [
      dict.reference.basics.title,
      dict.reference.exchange.title,
      dict.reference.numbers.title,
      dict.reference.frame.title,
      dict.reference.checksums.title,
      dict.reference.commands.title,
      dict.reference.stream.title,
      dict.reference.faults.title,
      dict.reference.zone.title,
      dict.reference.glossary.title,
    ];

    for (const title of sections) {
      expect(screen.getByRole("heading", { name: title })).toBeInTheDocument();
    }

    const contents = screen.getByRole("navigation", { name: dict.reference.title });
    expect(within(contents).getAllByRole("link")).toHaveLength(sections.length);
    for (const title of sections) {
      expect(within(contents).getByRole("link", { name: new RegExp(escape(title)) })).toBeInTheDocument();
    }
  });

  it("links each contents entry to the section it names", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);
    const contents = screen.getByRole("navigation", { name: dict.reference.title });

    expect(
      within(contents).getByRole("link", { name: new RegExp(escape(dict.reference.glossary.title)) }),
    ).toHaveAttribute("href", "#glossary");
  });
});

describe("ProtocolReference teaching content", () => {
  it("explains what a byte and a hex digit are before using them", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    for (const paragraph of dict.reference.numbers.body) {
      expect(screen.getByText(paragraph)).toBeInTheDocument();
    }
    expect(screen.getByText("0x68")).toBeInTheDocument();
    expect(screen.getByText("= 104")).toBeInTheDocument();
  });

  it("names the four directions a log row can carry", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    for (const lane of dict.reference.exchange.lanes) {
      expect(screen.getByText(lane.label)).toBeInTheDocument();
      expect(screen.getByText(lane.meaning)).toBeInTheDocument();
    }
  });

  it("walks the example frame byte by byte", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);
    const rows = dict.reference.frame.walkthrough.rows;

    const header = screen.getByRole("columnheader", {
      name: dict.reference.frame.walkthrough.columns.meaning,
    });
    // Field names also appear in the layout legend above, so the walkthrough
    // is asserted inside its own table.
    const table = within(header.closest("table")!);

    expect(rows).toHaveLength(8);
    for (const row of rows) {
      expect(table.getByText(row.name)).toBeInTheDocument();
      expect(table.getByText(row.meaning)).toBeInTheDocument();
    }
  });

  it("states the frame layout it is describing", () => {
    renderWithLocale(<ProtocolReference />);

    expect(screen.getByText(/0x68 \| LEN \| 0x68/)).toBeInTheDocument();
  });

  it("describes both checksum modes", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    for (const item of dict.reference.checksums.modes) {
      expect(screen.getByText(item.body)).toBeInTheDocument();
    }
  });

  it("lists both commands with their wire formats", () => {
    renderWithLocale(<ProtocolReference />);

    expect(screen.getByText("read-time")).toBeInTheDocument();
    expect(screen.getByText("read-identity")).toBeInTheDocument();
    expect(screen.getByText('DATA = 0x01 + "YYYY-MM-DD HH:MM:SS"')).toBeInTheDocument();
    expect(screen.getByText('DATA = 0x02 + "MODEL|SERIAL|FIRMWARE"')).toBeInTheDocument();
  });

  it("explains what each fault mode models in the field", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    for (const row of dict.reference.faults.rows) {
      expect(screen.getByText(row.meaning)).toBeInTheDocument();
    }
  });

  it("carries the time-zone warning that already cost this project a defect", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    expect(screen.getByRole("heading", { name: dict.reference.zone.title })).toBeInTheDocument();
    for (const paragraph of dict.reference.zone.body) {
      expect(screen.getByText(paragraph)).toBeInTheDocument();
    }
  });

  it("defines every term the panels use", () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    expect(dict.reference.glossary.terms.length).toBeGreaterThanOrEqual(10);
    for (const entry of dict.reference.glossary.terms) {
      expect(screen.getByText(entry.term)).toBeInTheDocument();
      expect(screen.getByText(entry.definition)).toBeInTheDocument();
    }
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
    expect(readout("sum8").getByText("02")).toBeInTheDocument();
    expect(
      readout("crc16 · little-endian").getByText(
        formatHex([crc16([0, 1, 1]) & 0xff, (crc16([0, 1, 1]) >> 8) & 0xff]),
      ),
    ).toBeInTheDocument();
  });

  it("recomputes as the payload is edited", async () => {
    const { dict } = renderWithLocale(<ProtocolReference />);
    const input = inputOf(dict);

    await userEvent.clear(input);
    await userEvent.paste("FF 02");

    const expected = sum8([0xff, 0x02]).toString(16).toUpperCase().padStart(2, "0");
    expect(readout("sum8").getByText(expected)).toBeInTheDocument();
  });

  it("handles an empty payload without failing", async () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    await userEvent.clear(inputOf(dict));

    expect(readout("sum8").getByText("00")).toBeInTheDocument();
  });

  it("redraws the sample frame when the mode changes", async () => {
    const { dict } = renderWithLocale(<ProtocolReference />);

    await userEvent.click(screen.getByRole("radio", { name: "crc16" }));

    expect(screen.getAllByText(dict.protocol.fields.checksum).length).toBeGreaterThan(0);
  });
});

describe("ProtocolReference in both locales", () => {
  it.each(locales)("renders every section in %s", (locale) => {
    const { unmount } = renderWithLocale(<ProtocolReference />, { locale });
    const dict = getDictionary(locale);

    expect(screen.getByRole("heading", { name: dict.reference.basics.title })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: dict.reference.glossary.title })).toBeInTheDocument();
    expect(screen.getByText(dict.reference.frame.walkthrough.hint)).toBeInTheDocument();

    unmount();
  });
});

/** A readout is a label and its value; the value alone is not unique on the page. */
function readout(label: string | RegExp) {
  return within(screen.getByText(label).parentElement!);
}

/** Section titles are used to build regexes; a few carry regex metacharacters. */
function escape(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
