import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { Badge } from "./badge";
import { Button } from "./button";
import {
  Field,
  Input,
  SegmentedControl,
  Select,
  SliderRow,
  Switch,
  TextArea,
  ToggleRow,
} from "./controls";
import { DefRow, Panel, PanelBody, PanelHeader, Stat } from "./panel";
import { StatusDot } from "./status-dot";

describe("Button", () => {
  it("renders its label and reacts to a click", async () => {
    const onClick = jest.fn();
    render(<Button onClick={onClick}>Запустить</Button>);

    await userEvent.click(screen.getByRole("button", { name: "Запустить" }));

    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("does not fire while disabled", async () => {
    const onClick = jest.fn();
    render(
      <Button disabled onClick={onClick}>
        Стоп
      </Button>,
    );

    await userEvent.click(screen.getByRole("button", { name: "Стоп" }));

    expect(onClick).not.toHaveBeenCalled();
  });

  it.each(["default", "outline", "subtle", "ghost", "danger", "success"] as const)(
    "renders the %s variant",
    (variant) => {
      render(<Button variant={variant}>x</Button>);

      expect(screen.getByRole("button")).toBeInTheDocument();
    },
  );

  it("renders as a child element when asked", () => {
    render(
      <Button asChild>
        <a href="/ru">Обзор</a>
      </Button>,
    );

    expect(screen.getByRole("link", { name: "Обзор" })).toHaveAttribute("href", "/ru");
  });

  it("merges an extra class rather than dropping it", () => {
    render(<Button className="custom-class">x</Button>);

    expect(screen.getByRole("button")).toHaveClass("custom-class");
  });
});

describe("Badge", () => {
  it("renders its content", () => {
    render(<Badge tone="success">норма</Badge>);

    expect(screen.getByText("норма")).toBeInTheDocument();
  });

  it.each(["neutral", "primary", "success", "warning", "danger", "tx", "rx", "err", "sys"] as const)(
    "renders the %s tone",
    (tone) => {
      render(<Badge tone={tone}>{tone}</Badge>);

      expect(screen.getByText(tone)).toBeInTheDocument();
    },
  );

  it("switches to monospace when asked", () => {
    render(<Badge mono>68 03</Badge>);

    expect(screen.getByText("68 03")).toHaveClass("font-mono");
  });
});

describe("StatusDot", () => {
  it.each(["online", "degraded", "offline", "idle", "unknown"] as const)(
    "renders the %s tone",
    (tone) => {
      const { container } = render(<StatusDot tone={tone} />);

      expect(container.firstChild).toBeInTheDocument();
    },
  );

  it("pulses only when told to", () => {
    const { container: still } = render(<StatusDot tone="online" />);
    const { container: pulsing } = render(<StatusDot tone="online" pulse />);

    expect(still.querySelector(".pulse-dot")).toBeNull();
    expect(pulsing.querySelector(".pulse-dot")).not.toBeNull();
  });
});

describe("Panel", () => {
  it("renders a header, hint and body", () => {
    render(
      <Panel>
        <PanelHeader title="Монитор обмена" hint="Живой поток" actions={<Button>x</Button>} />
        <PanelBody>содержимое</PanelBody>
      </Panel>,
    );

    expect(screen.getByRole("heading", { name: "Монитор обмена" })).toBeInTheDocument();
    expect(screen.getByText("Живой поток")).toBeInTheDocument();
    expect(screen.getByText("содержимое")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "x" })).toBeInTheDocument();
  });

  it("omits the hint and actions when not supplied", () => {
    render(<PanelHeader title="Счётчики" />);

    expect(screen.getByRole("heading", { name: "Счётчики" })).toBeInTheDocument();
  });

  it("renders an icon when given one", () => {
    render(<PanelHeader title="t" icon={<span data-testid="icon" />} />);

    expect(screen.getByTestId("icon")).toBeInTheDocument();
  });
});

describe("Stat", () => {
  it("renders a label, a value and a hint", () => {
    render(<Stat label="Успешных чтений" value={42} hint="за сессию" />);

    expect(screen.getByText("Успешных чтений")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
    expect(screen.getByText("за сессию")).toBeInTheDocument();
  });

  it("uses tabular monospace for protocol values", () => {
    render(<Stat label="l" value="68 03" mono />);

    expect(screen.getByText("68 03")).toHaveClass("font-mono", "tabular");
  });

  it.each(["default", "muted", "success", "warning", "danger", "primary"] as const)(
    "renders the %s tone",
    (tone) => {
      render(<Stat label="l" value="v" tone={tone} />);

      expect(screen.getByText("v")).toBeInTheDocument();
    },
  );
});

describe("DefRow", () => {
  it("pairs a label with a value", () => {
    render(<DefRow label="Переподключений" value={3} />);

    expect(screen.getByText("Переподключений")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });
});

describe("Field", () => {
  it("associates its label with the control", () => {
    render(
      <Field label="Адрес" hint="0..255" htmlFor="addr">
        <Input id="addr" defaultValue="1" />
      </Field>,
    );

    expect(screen.getByLabelText("Адрес")).toHaveValue("1");
    expect(screen.getByText("0..255")).toBeInTheDocument();
  });
});

describe("Input and TextArea", () => {
  it("accepts typing", async () => {
    render(<Input aria-label="hex" defaultValue="" />);

    await userEvent.type(screen.getByLabelText("hex"), "68 03");

    expect(screen.getByLabelText("hex")).toHaveValue("68 03");
  });

  it("accepts multiline input", async () => {
    render(<TextArea aria-label="frame" defaultValue="" />);

    await userEvent.type(screen.getByLabelText("frame"), "68");

    expect(screen.getByLabelText("frame")).toHaveValue("68");
  });
});

describe("Select", () => {
  it("changes the selected option", async () => {
    const onChange = jest.fn();
    render(
      <Select aria-label="режим" defaultValue="sum" onChange={onChange}>
        <option value="sum">sum</option>
        <option value="crc16">crc16</option>
      </Select>,
    );

    await userEvent.selectOptions(screen.getByLabelText("режим"), "crc16");

    expect(screen.getByLabelText("режим")).toHaveValue("crc16");
    expect(onChange).toHaveBeenCalled();
  });
});

describe("Switch", () => {
  it("toggles", async () => {
    const onCheckedChange = jest.fn();
    render(<Switch aria-label="молчание" checked={false} onCheckedChange={onCheckedChange} />);

    await userEvent.click(screen.getByRole("switch"));

    expect(onCheckedChange).toHaveBeenCalledWith(true);
  });
});

describe("ToggleRow", () => {
  it("labels the switch and reports a change", async () => {
    const onCheckedChange = jest.fn();
    render(
      <ToggleRow
        label="Молчание"
        hint="Прибор не отвечает вовсе."
        checked={false}
        onCheckedChange={onCheckedChange}
      />,
    );

    expect(screen.getByText("Прибор не отвечает вовсе.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("switch", { name: "Молчание" }));

    expect(onCheckedChange).toHaveBeenCalledWith(true);
  });

  it("does not toggle while disabled", async () => {
    const onCheckedChange = jest.fn();
    render(
      <ToggleRow label="Молчание" checked={false} onCheckedChange={onCheckedChange} disabled />,
    );

    await userEvent.click(screen.getByRole("switch"));

    expect(onCheckedChange).not.toHaveBeenCalled();
  });
});

describe("SliderRow", () => {
  it("shows the formatted value and reports a change", () => {
    const onChange = jest.fn();
    render(
      <SliderRow
        label="Задержка ответа"
        value={100}
        min={0}
        max={1000}
        step={50}
        format={(value) => `${value} ms`}
        onChange={onChange}
      />,
    );

    expect(screen.getByText("100 ms")).toBeInTheDocument();

    const slider = screen.getByRole("slider", { name: "Задержка ответа" });
    fireChange(slider as HTMLInputElement, "250");

    expect(onChange).toHaveBeenCalledWith(250);
  });

  it("renders a hint when given one", () => {
    render(
      <SliderRow
        label="l"
        hint="подсказка"
        value={0}
        min={0}
        max={1}
        step={1}
        format={String}
        onChange={jest.fn()}
      />,
    );

    expect(screen.getByText("подсказка")).toBeInTheDocument();
  });
});

describe("SegmentedControl", () => {
  it("marks the active option and reports a change", async () => {
    const onChange = jest.fn();
    render(
      <SegmentedControl
        value="sum"
        options={[
          { value: "sum", label: "sum" },
          { value: "crc16", label: "crc16" },
        ]}
        onChange={onChange}
      />,
    );

    expect(screen.getByRole("radio", { name: "sum" })).toHaveAttribute("aria-checked", "true");
    expect(screen.getByRole("radio", { name: "crc16" })).toHaveAttribute("aria-checked", "false");

    await userEvent.click(screen.getByRole("radio", { name: "crc16" }));

    expect(onChange).toHaveBeenCalledWith("crc16");
  });
});

/**
 * `userEvent` cannot drag a range input, so a change event is dispatched
 * directly — the assertion is about the handler, not about the pointer.
 */
function fireChange(element: HTMLInputElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(
    window.HTMLInputElement.prototype,
    "value",
  )?.set;
  setter?.call(element, value);
  element.dispatchEvent(new Event("change", { bubbles: true }));
}
