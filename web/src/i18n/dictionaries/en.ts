import type { Dictionary } from "../index";

/**
 * English dictionary.
 *
 * Checked against the shape of `ru` (the canonical locale): a key added there
 * and not here is a compile error, which is what keeps the second locale from
 * silently rotting behind the first.
 */
export const en: Dictionary = {
  locale: "en",

  meta: {
    title: "FT1.2 / TTR20 — protocol bench",
    description:
      "Engineering console for the device emulator, polling gateway and FT1.2 frame analysis.",
  },

  nav: {
    overview: "Overview",
    protocol: "Frames",
    monitor: "Monitor",
    emulator: "Emulator",
    gateway: "Gateway",
    reference: "Reference",
    sectionWork: "Work",
    sectionControl: "Control",
    sectionLearn: "Learn",
  },

  shell: {
    brand: "FT1.2 / TTR20",
    brandSub: "protocol bench",
    benchMode: "Standalone bench",
    benchModeHint:
      "Data comes from the built-in device model. No backend needed — this is the teaching and demo mode.",
    statusStrip: "Bench status",
    language: "Language",
  },

  source: {
    label: "Data source",
    bench: "Bench",
    live: "Live stack",
    benchHint:
      "Everything is computed by the built-in device model, in this browser. No backend needed: this is the teaching and demo mode.",
    liveHint:
      "The console reads a running Go stack over the HTTP API. Every number on screen comes from the real gateway and the real device.",
    switchTo: "Switch source",
    linkReady: "Connected",
    linkConnecting: "Connecting",
    linkUnreachable: "API unreachable",
    unreachableTitle: "The Go API is not answering",
    unreachableBody:
      "The console calls the HTTP API at /upstream, which proxies through to ft12-api. That process does not appear to be running.",
    unreachableCommand:
      "Start the three processes — each in its own terminal — then come back to this page:",
    upstreamError: "The gateway returned an error",
    readOnly: "Read-only",
    readOnlyHint:
      "Interval, thresholds and the availability policy come from the gateway process's own configuration. What you see here are its actual values — the console does not change them.",
    target: "Polled device",
    benchOnly: "Bench only",
    benchOnlyHint:
      "Clock offset and drift are produced by the device model in this browser. The Go emulator has no such setting, so these controls are unavailable on the live source.",
    liveFaultsHint:
      "These switches change the fault mode of the real Go emulator, over the API.",
    emulatorUnavailable: "The emulator is not answering — the fault mode cannot be changed.",
    identityLive:
      "The nameplate is set by the emulator itself; on the live source the console only shows what was read.",
    noReset: "Gateway counters are cleared only by restarting it.",
  },

  common: {
    no: "no",
    none: "—",
    unknown: "unknown",
    copy: "Copy",
    copied: "Copied",
    clear: "Clear",
    reset: "Reset",
    start: "Start",
    stop: "Stop",
    running: "running",
    stopped: "stopped",
    connected: "connected",
    disconnected: "disconnected",
    bytes: "bytes",
    ms: "ms",
    perDay: "/day",
    of: "of",
    total: "total",
    request: "request",
    response: "response",
  },

  units: {
    ms: "ms",
    s: "s",
    m: "m",
    h: "h",
    perDay: "/day",
  },

  states: {
    online: "online",
    degraded: "degraded",
    offline: "offline",
    unknown: "unknown",
    ok: "ok",
    warn: "warning",
    critical: "critical",
  },

  directions: {
    tx: "TX",
    rx: "RX",
    err: "ERR",
    system: "SYS",
    txFull: "transmit",
    rxFull: "receive",
    errFull: "error",
    systemFull: "event",
  },

  overview: {
    title: "Bench overview",
    subtitle: "Device, gateway and the current polling cycle",
    deviceTime: "Device time",
    serverTime: "Server time",
    skew: "Clock skew",
    skewHint:
      "Difference between device time and the reference, compensated for half the round trip.",
    drift: "Drift rate",
    driftHint: "Linear regression over the sample window.",
    roundTrip: "Round trip",
    availability: "Availability",
    successfulReads: "Successful reads",
    failedReads: "Failed reads",
    reconnects: "Reconnects",
    retries: "In-session retries",
    protocolErrors: "Protocol errors",
    nextPoll: "Next poll",
    schedule: "Schedule",
    lastCycle: "Last cycle",
    eventRate: "Frame rate",
    eventRateHint: "Frames over the last minute, per second.",
    model: "Model",
    serial: "Serial",
    firmware: "Firmware",
    emulator: "Device emulator",
    gateway: "Polling gateway",
    target: "Target",
    sessions: "Sessions",
    checksumMode: "Checksum",
  },

  fleet: {
    title: "Device fleet",
    hint: "Every device this gateway polls, and how each one is doing.",
    device: "Device",
    target: "Address",
    state: "State",
    clock: "Clock",
    skew: "Skew",
    availability: "Availability",
    samples: "Samples",
    worst: "Largest skew",
    polling: "Polling",
    idle: "Not polling",
    empty: "The gateway reports no devices.",
    single:
      "The gateway was started without a device list and polls a single device. The list comes from an inventory file.",
  },

  protocol: {
    title: "Frame analyzer",
    subtitle: "Byte-level FT1.2 decode with checksum verification",
    input: "Frame as hex",
    inputHint: "Any separators: spaces, commas, newlines. A 0x prefix is fine.",
    mode: "Checksum mode",
    builder: "Builder",
    builderHint: "Assemble a frame from its fields — it lands in the decode box.",
    control: "CONTROL",
    address: "ADDRESS",
    command: "Command",
    payload: "Payload",
    responseBit: "Response bit (0x80)",
    insert: "Insert",
    samples: "Ready-made samples",
    layout: "Frame layout",
    byteMap: "Byte map",
    offset: "Offset",
    hex: "HEX",
    bin: "BIN",
    ascii: "ASCII",
    field: "Field",
    valid: "Frame is valid",
    invalid: "Frame has errors",
    empty: "Enter a frame or build one",
    expected: "expected",
    actual: "got",
    computed: "Computed checksum",
    payloadSpan: "The checksum covers CONTROL + ADDRESS + DATA",
    frameLength: "Frame length",
    payloadLength: "Payload length",
    direction: "Direction",
    streamTitle: "Stream scan",
    streamHint:
      "Paste an arbitrary byte stream: the parser extracts frames, drops noise and shows the remainder.",
    copyFrame: "Copy frame",
    streamFrames: "Frames found",
    streamRest: "Buffer remainder",
    fields: {
      start: "Start byte",
      length: "Length",
      startRepeat: "Start repeat",
      control: "Control",
      address: "Address",
      data: "Data",
      checksum: "Checksum",
      end: "End byte",
      trailing: "Trailing",
    },
    errors: {
      tooShort: "Frame is shorter than the minimum",
      invalidStart: "Invalid start byte",
      invalidStartRepeat: "Invalid repeated start byte",
      invalidLength: "Invalid payload length",
      invalidChecksum: "Checksum mismatch",
      invalidEnd: "Invalid end byte",
      tooLarge: "Frame exceeds the maximum size",
      emptyPayload: "Empty command payload",
      invalidTimestamp: "Invalid timestamp",
      invalidIdentity: "Invalid device identity",
      invalidLengthPayload: "Invalid payload length",
    },
    commandInfo: {
      readTime: "Read the device date and time",
      readIdentity: "Read model, serial number and firmware",
    },
  },

  events: {
    pollingStarted: "Polling started",
    pollingStopped: "Polling stopped",
    deviceStateChanged: "Device state changed",
    clockStateChanged: "Clock state changed",
    reconnected: "Connection re-established",
    errors: {
      invalidChecksum: "Checksum did not match — frame discarded",
      noResponse: "The device did not answer in time",
      timeout: "Response timeout expired",
      connectionClosed: "The device closed the connection",
    },
  },

  monitor: {
    title: "Exchange monitor",
    subtitle: "Live frame stream between gateway and device",
    search: "Search hex, command, address",
    autoscroll: "Auto-scroll",
    clearLog: "Clear log",
    empty: "No frames yet",
    emptyHint: "Start polling to see the exchange.",
    time: "Time",
    dir: "Dir",
    direction: "Exchange direction",
    source: "Source",
    commandColumn: "Command",
    latency: "Latency",
    raw: "Frame",
    detail: "Frame decode",
    detailHint: "Pick a row from the log",
    exchange: "Exchange diagram",
    exchangeHint: "Recent polling cycles: request, response and the delay between them.",
    counters: "Counters",
    frames: "frames",
  },

  emulator: {
    title: "Device emulator",
    subtitle: "TTR20 model: responses, faults and clock behaviour",
    faults: "Fault modes",
    faultsHint: "Every switch reproduces a real line or device fault.",
    responseDelay: "Response delay",
    responseDelayHint: "The device answers late — exercises gateway timeouts.",
    badChecksum: "Checksum corruption",
    badChecksumHint: "Share of answers with a broken checksum. Exercises retries.",
    fragment: "Frame fragmentation",
    fragmentHint: "The answer arrives in pieces — exercises the streaming parser.",
    fragmentDelay: "Delay between fragments",
    noResponse: "Silence",
    noResponseHint: "The device does not answer at all.",
    closeAfterRequest: "Drop after request",
    closeAfterRequestHint: "The device closes the connection right after a request.",
    clock: "Device clock",
    clockHint: "Offset and drift — what skew monitoring is looking for.",
    clockOffset: "Constant offset",
    clockDrift: "Drift",
    identity: "Identity",
    presets: "Scenarios",
    presetHealthy: "Healthy device",
    presetNoisyLine: "Noisy line",
    presetSlowDevice: "Slow device",
    presetDriftingClock: "Drifting clock",
    presetDeadDevice: "Dead device",
    presetHint: "One-click states for demonstration.",
    activeFaults: "Active faults",
  },

  gateway: {
    title: "Polling gateway",
    subtitle: "Schedule, retries, device state",
    polling: "Polling",
    pollingRunning: "Polling is running",
    pollingStopped: "Polling is stopped",
    scheduleMode: "Schedule mode",
    scheduleInterval: "By interval",
    scheduleAligned: "By calendar",
    scheduleIntervalHint: "Counted from the moment of connection.",
    scheduleAlignedHint:
      "Polls land on calendar boundaries — for example exactly on the fifth second of every minute.",
    interval: "Interval",
    offset: "Offset inside the interval",
    requestTimeout: "Request timeout",
    retryAttempts: "Retries on a frame error",
    retryHint:
      "A frame error is retried on the same connection. The link is dropped only on a transport failure.",
    clockThresholds: "Clock skew thresholds",
    warnThreshold: "Warning",
    criticalThreshold: "Critical",
    healthPolicy: "Device state",
    healthPolicyHint:
      "How many consecutive failed polls degrade the device and then take it offline, and how many successful ones bring it back.",
    degradeAfter: "Degrade after",
    offlineAfter: "Offline after",
    recoverAfter: "Recover after",
    failuresShort: "fail",
    successesShort: "ok",
    deviceState: "Device state",
    connection: "Connection",
    timeline: "Schedule timeline",
    timelineHint: "Ticks are poll moments. Aligned mode keeps its phase.",
    skewChart: "Skew history",
    skewChartHint: "One sample per successful cycle; the line is the window median.",
  },

  reference: {
    title: "Reference",
    subtitle: "From zero: what happens here, how to read a frame and what every symbol means",

    basics: {
      title: "Start here",
      lead: "If this is your first time at the bench, read this section — everything after it will make sense.",
      body: [
        "An electricity meter sits on site and counts consumption. It has no network of its own: beside it stands a TTR20 teleport adapter that translates network requests onto the meter serial line and back.",
        "A gateway program periodically asks the meter for its current time, and the meter answers. That exchange is everything this bench does.",
        "Why ask for the time? The meter records consumption in half-hourly and hourly slices and stamps each record from its own clock. If that clock drifts, the whole archive shifts with it: consumption that happened at 09:00 lands in the 08:30 slice. That is why clock skew is tracked separately and continuously — it is the headline figure on this bench.",
      ],
    },

    exchange: {
      title: "How an exchange works",
      body: [
        "The line is master and slave. The gateway always speaks first: it sends a request and waits. The device stays silent and answers only when addressed.",
        "One request, one answer. The next request does not go out until the answer arrives or the timeout expires, which is why log rows come in pairs — a transmit followed immediately by a receive.",
      ],
      lanes: [
        { label: "TX", meaning: "The gateway asks — a frame went out on the line" },
        { label: "RX", meaning: "The device answered — a frame came back" },
        { label: "ERR", meaning: "No answer, or one that arrived damaged" },
        { label: "SYS", meaning: "The bench reporting a state change of its own" },
      ],
    },

    numbers: {
      title: "Bytes and hexadecimal",
      body: [
        "What travels on the line is not letters but bytes. A byte is a number from 0 to 255 — eight binary digits, or bits.",
        "Writing bytes in decimal is awkward, so hexadecimal is used instead: sixteen digits, 0–9 and A–F, where A means 10, B means 11 and so on up to F, which is 15. One byte is exactly two of those digits.",
        "The 0x prefix means that a hexadecimal number follows. So 0x68 is the familiar 104, 0x16 is 22 and 0xFF is 255. In the log and the analyzer, bytes appear as digit pairs with no prefix: 68 03 68 and so on.",
        "A bit is one binary digit of a byte. When something is described as having bit 0x80 set, that is the highest of the eight: it adds 128 to the value. That is the mark the device uses to say a frame is an answer rather than a request.",
      ],
    },

    frame: {
      title: "Frame format",
      lead: "A frame is one complete message: a wrapper of housekeeping bytes with the useful data inside it.",
      body: [
        "Frames vary in length, so the receiver has to work out for itself where a message began and where it ended. The start bytes, the length byte and the end byte are what let it.",
      ],
      walkthrough: {
        title: "Byte by byte",
        hint: "A read-time request in sum mode, taken apart: 68 03 68 00 01 01 02 16",
        columns: { byte: "Byte", name: "Field", meaning: "What it means" },
        rows: [
          {
            byte: "68",
            name: "Start byte",
            meaning: "The frame begins. Anything that arrived earlier is noise, and the receiver discards it.",
          },
          {
            byte: "03",
            name: "Length",
            meaning:
              "How many bytes CONTROL, ADDRESS and DATA occupy together. Three here: one control byte, one address byte and one byte of data.",
          },
          {
            byte: "68",
            name: "Repeated start byte",
            meaning:
              "The same byte again. It marks a variable-length frame and guards against a chance match in the stream.",
          },
          {
            byte: "00",
            name: "CONTROL",
            meaning:
              "The housekeeping byte. 0x00 means a request from the gateway; in an answer the device sets bit 0x80 and the byte becomes 0x80.",
          },
          {
            byte: "01",
            name: "ADDRESS",
            meaning:
              "The address of the device on the line. Several devices can share one line, and each answers only to its own address.",
          },
          {
            byte: "01",
            name: "DATA",
            meaning: "The payload. Its first byte is the command id: 0x01 means read the time.",
          },
          {
            byte: "02",
            name: "Checksum",
            meaning:
              "A check number computed over CONTROL, ADDRESS and DATA. A mismatch means the frame was damaged on the way.",
          },
          {
            byte: "16",
            name: "End byte",
            meaning: "The frame ends. After it, the message counts as read in full.",
          },
        ],
      },
    },

    checksums: {
      title: "Checksums",
      body: [
        "A line is never perfect: interference flips a bit and a byte arrives as a different byte. Such a frame looks perfectly ordinary, so a check number is carried inside it.",
        "The sender computes a checksum over the contents of the frame and writes it in. The receiver computes it again over the same bytes and compares. A mismatch means the frame is discarded and the gateway asks again.",
        "The checksum covers CONTROL, ADDRESS and DATA only. The start bytes, the length byte and the end byte are outside it — their position in the frame already checks them.",
      ],
      modes: [
        {
          name: "sum",
          body:
            "A plain addition of every byte modulo 256 — the remainder of the total divided by 256. One byte. It catches single corruptions, but two bytes swapped would slip past it.",
        },
        {
          name: "crc16",
          body:
            "CRC-16/Modbus, a cyclic redundancy check: polynomial 0xA001, initial value 0xFFFF. Two bytes, sent least significant byte first. It catches considerably more kinds of damage than a plain sum.",
        },
      ],
      calculator: "Calculator",
      calculatorHint: "Enter the payload bytes — both checksums are recomputed as you type.",
      calculatorInput: "CONTROL + ADDRESS + DATA bytes",
    },

    commands: {
      title: "Commands",
      body: [
        "A command is what the gateway asks the device for. Its id is always the first byte of DATA, and whatever follows in that field is the substance of the answer.",
        "An answer carries the same command id as the request. The only thing separating them is bit 0x80 in the control byte.",
      ],
      columns: {
        id: "ID",
        name: "Name",
        purpose: "Purpose",
        request: "Request",
        response: "Response",
      },
    },

    stream: {
      title: "Streams and resynchronisation",
      body: [
        "A line carries a continuous stream of bytes, not messages. One frame can arrive in two pieces and two frames in one; that is normal behaviour, not a fault.",
        "So the receiver does not read one frame at a time. It accumulates bytes and extracts frames itself. When it meets rubbish mid-stream it looks for the next start byte and carries on from there — that is resynchronisation.",
        "You can try it in the analyzer: paste several frames in a row mixed with stray bytes and watch what it extracts and what it leaves in the remainder.",
      ],
    },

    faults: {
      title: "What the fault modes model",
      body: [
        "The emulator panel reproduces failures that really happen on a line. Each one exercises a different part of the gateway.",
      ],
      rows: [
        {
          name: "Response delay",
          meaning: "The device is busy, or the line is slow. Exercises the request timeout.",
        },
        {
          name: "Checksum corruption",
          meaning:
            "Interference on the line. Exercises whether the gateway retries instead of dropping the connection.",
        },
        {
          name: "Frame fragmentation",
          meaning: "The answer arrives in pieces. Exercises the streaming parser.",
        },
        {
          name: "Silence",
          meaning: "The device is powered down or off the line. Exercises timeouts and the fall to offline.",
        },
        {
          name: "Drop after request",
          meaning: "A broken link or a rebooting adapter. Exercises reconnection.",
        },
        {
          name: "Clock offset and drift",
          meaning:
            "The device clock wanders. Exercises skew monitoring — the one fault where the protocol works flawlessly and the data is wrong anyway.",
        },
      ],
    },

    zone: {
      title: "The time-zone trap",
      body: [
        "The timestamp travels as the text 2026-09-02 22:41:15 — with no zone attached.",
        "That means sender and receiver must agree on which zone it is written in. If the device writes local time and the program reads it as UTC, every reading is offset by exactly the difference between them — and the monitor reports a skew on a device whose clock is fine.",
        "This project made that mistake once and the tests found it. It is why encoding and decoding here take the zone explicitly rather than trusting a default.",
      ],
    },

    glossary: {
      title: "Glossary",
      hint: "Words that appear on the bench panels.",
      terms: [
        { term: "Frame", definition: "One complete message, from the start byte to the end byte." },
        { term: "Poll", definition: "The gateway addressing the device on a schedule." },
        {
          term: "Cycle",
          definition: "One full exchange: request sent, answer parsed. Retries belong to the same cycle.",
        },
        {
          term: "Round trip",
          definition:
            "The time from sending a request to receiving the answer. Half of it is the correction applied when computing clock skew.",
        },
        {
          term: "Clock skew",
          definition:
            "The difference between device time and the reference, corrected for the line. The sign matters: plus means the device runs fast, minus means it runs slow.",
        },
        {
          term: "Drift",
          definition: "The rate at which skew grows. Estimated over a window of samples and quoted per day.",
        },
        {
          term: "Median",
          definition:
            "The middle value of a sample set. Clock state is judged on it so that a single outlier does not raise an alarm.",
        },
        { term: "Availability", definition: "The share of successful polls in the most recent window." },
        {
          term: "Hysteresis",
          definition:
            "Different thresholds for entering and leaving a state. It stops a device flapping between online and offline on an unstable line.",
        },
        {
          term: "Retry",
          definition:
            "Another attempt on the same connection after a frame error. A dropped link is a different case and needs a reconnection.",
        },
        {
          term: "Aligned schedule",
          definition:
            "Polls land on calendar boundaries — the fifth second of every minute, say — rather than every five seconds counted from startup.",
        },
      ],
    },

    errorsTitle: "Decode errors",
    errorsHint: "The same codes the project Go core returns.",
  },

};
