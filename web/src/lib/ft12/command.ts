/**
 * Command layer, mirroring `internal/protocol/command`.
 *
 * The registry is deliberately data rather than a switch: the Go side models
 * commands the same way, and the console renders the command list straight
 * from it, so teaching the bench a new command is one entry here.
 */

export const RESPONSE_BIT = 0x80;

export const READ_TIME = 0x01;
export const READ_IDENTITY = 0x02;

export const READ_TIME_LAYOUT = "YYYY-MM-DD HH:MM:SS";
export const IDENTITY_SEPARATOR = "|";

export interface CommandDescriptor {
  id: number;
  name: string;
  /** Dictionary key for the human description — the UI is bilingual. */
  descriptionKey: "readTime" | "readIdentity";
}

export const COMMANDS: readonly CommandDescriptor[] = [
  { id: READ_TIME, name: "read-time", descriptionKey: "readTime" },
  { id: READ_IDENTITY, name: "read-identity", descriptionKey: "readIdentity" },
];

export function commandById(id: number): CommandDescriptor | undefined {
  return COMMANDS.find((command) => command.id === id);
}

export function commandName(id: number): string {
  return commandById(id)?.name ?? `unknown-0x${id.toString(16).toUpperCase().padStart(2, "0")}`;
}

export type PayloadKind = "readTimeRequest" | "readTimeResponse" | "readIdentityRequest" | "readIdentityResponse" | "ack" | "unknown";

export interface ParsedPayload {
  kind: PayloadKind;
  commandId?: number;
  commandName: string;
  /** Rendered value — a timestamp, an identity triple, an ACK body. */
  value?: string;
  fields?: { label: string; value: string }[];
  error?: "emptyPayload" | "invalidTimestamp" | "invalidIdentity" | "invalidLength";
}

const asciiOf = (bytes: readonly number[]): string =>
  bytes.map((byte) => String.fromCharCode(byte)).join("");

const TIMESTAMP_PATTERN = /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/;

/**
 * Interprets a frame's DATA field.
 *
 * `isResponse` comes from the control byte rather than from the payload: the
 * emulator answers with bit 0x80 set, and a request and its answer carry the
 * same command id, so the direction is the only thing that separates them.
 */
export function parsePayload(data: readonly number[], isResponse: boolean): ParsedPayload {
  if (data.length === 0) {
    return { kind: "unknown", commandName: "—", error: "emptyPayload" };
  }

  const id = data[0]!;
  const name = commandName(id);
  const body = data.slice(1);

  if (id === READ_TIME) {
    if (!isResponse) {
      return { kind: "readTimeRequest", commandId: id, commandName: name };
    }
    const text = asciiOf(body);
    if (body.length !== READ_TIME_LAYOUT.length) {
      return {
        kind: "readTimeResponse",
        commandId: id,
        commandName: name,
        value: text,
        error: "invalidLength",
      };
    }
    return {
      kind: "readTimeResponse",
      commandId: id,
      commandName: name,
      value: text,
      error: TIMESTAMP_PATTERN.test(text) ? undefined : "invalidTimestamp",
    };
  }

  if (id === READ_IDENTITY) {
    if (!isResponse) {
      return { kind: "readIdentityRequest", commandId: id, commandName: name };
    }
    const text = asciiOf(body);
    const parts = text.split(IDENTITY_SEPARATOR);
    if (parts.length !== 3 || parts.some((part) => part.trim() === "")) {
      return {
        kind: "readIdentityResponse",
        commandId: id,
        commandName: name,
        value: text,
        error: "invalidIdentity",
      };
    }
    return {
      kind: "readIdentityResponse",
      commandId: id,
      commandName: name,
      value: text,
      fields: [
        { label: "model", value: parts[0]! },
        { label: "serial", value: parts[1]! },
        { label: "firmware", value: parts[2]! },
      ],
    };
  }

  const text = asciiOf(body);
  if (isResponse && text === "OK") {
    return { kind: "ack", commandId: id, commandName: name, value: text };
  }
  return { kind: "unknown", commandId: id, commandName: name, value: text || undefined };
}

export function buildReadTimeRequest(): number[] {
  return [READ_TIME];
}

export function buildReadIdentityRequest(): number[] {
  return [READ_IDENTITY];
}

/** Renders a Date the way the wire format does — naive local wall clock. */
export function formatDeviceTime(date: Date): string {
  const pad = (value: number, width = 2) => String(value).padStart(width, "0");
  return (
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ` +
    `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
  );
}

export function buildReadTimeResponse(date: Date): number[] {
  return [READ_TIME, ...[...formatDeviceTime(date)].map((char) => char.charCodeAt(0))];
}

export function buildReadIdentityResponse(model: string, serial: string, firmware: string): number[] {
  const body = [model, serial, firmware].join(IDENTITY_SEPARATOR);
  return [READ_IDENTITY, ...[...body].map((char) => char.charCodeAt(0))];
}
