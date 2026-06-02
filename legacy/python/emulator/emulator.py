import argparse
import asyncio
import datetime
import logging
import sys

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(message)s",
    handlers=[
        logging.FileHandler("emulator.log", encoding="utf-8"),
        logging.StreamHandler(sys.stdout),
    ],
)

START = 0x68
END = 0x16


def build_ft12_frame(payload: bytes, control: int = 0x53, address: int = 0x01) -> bytes:
    L = 1 + 1 + len(payload)
    frame = bytearray()
    frame.append(START)
    frame.append(L & 0xFF)
    frame.append(L & 0xFF)
    frame.append(START)
    frame.append(control & 0xFF)
    frame.append(address & 0xFF)
    frame.extend(payload)
    cs = sum(frame[4:]) & 0xFF
    frame.append(cs)
    frame.append(END)
    return bytes(frame)


def parse_ft12_frame(buf: bytes):
    if len(buf) < 6:
        raise ValueError("too short")
    if buf[0] != START or buf[3] != START:
        raise ValueError("invalid start bytes")
    L1 = buf[1]
    L2 = buf[2]
    if L1 != L2:
        raise ValueError("length mismatch")
    L = L1
    expected = 4 + L + 2
    if len(buf) != expected:
        raise ValueError(f"frame length {len(buf)} != expected {expected}")
    control = buf[4]
    address = buf[5]
    payload = buf[6 : 6 + (L - 2)]
    cs = buf[6 + (L - 2)]
    calc = sum(buf[4 : 6 + (L - 2)]) & 0xFF
    if cs != calc:
        raise ValueError(f"checksum mismatch {cs} != {calc}")
    if buf[-1] != END:
        raise ValueError("invalid end char")
    return {"control": control, "address": address, "payload": bytes(payload)}


def make_text_response() -> bytes:
    now = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    return f"TIME:{now}\n".encode("utf-8")


def make_ft12_time_payload() -> bytes:
    return f"TIME:{datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}".encode(
        "ascii"
    )


class UDPHandler:
    def __init__(self, emulator):
        self.emulator = emulator

    def connection_made(self, transport):
        self.transport = transport

    def datagram_received(self, data, addr):
        logging.info("UDP received from %s: %r", addr, data)
        try:
            if data and data[0] == START and len(data) >= 6:
                try:
                    info = parse_ft12_frame(data)
                    logging.info(
                        "Parsed FT1.2 request: control=0x%02X addr=0x%02X payload=%r",
                        info["control"],
                        info["address"],
                        info["payload"],
                    )
                    resp_payload = make_ft12_time_payload()
                    resp = build_ft12_frame(
                        resp_payload, control=0x73, address=info["address"]
                    )
                except Exception as e:
                    logging.warning(
                        "FT1.2 parse failed: %s - falling back to text response", e
                    )
                    resp = make_text_response()
            else:
                s = None
                try:
                    s = data.decode("utf-8", errors="ignore")
                except:
                    s = None
                if s and s.strip().upper() == "GETTIME":
                    resp = make_text_response()
                else:
                    resp = make_text_response()
            self.transport.sendto(resp, addr)
            logging.info("UDP sent to %s: %r", addr, resp)
        except Exception:
            logging.exception("UDP handler error")


class Emulator:
    def __init__(self, host: str, port: int, proto: str, mode: str):
        self.host = host
        self.port = port
        self.proto = proto.lower()
        self.mode = mode.lower()

    async def start(self):
        logging.info(
            "Starting emulator on %s:%d proto=%s mode=%s",
            self.host,
            self.port,
            self.proto,
            self.mode,
        )
        if self.proto == "tcp":
            server = await asyncio.start_server(
                self.handle_tcp, host=self.host, port=self.port
            )
            addrs = ", ".join(str(sock.getsockname()) for sock in server.sockets)
            logging.info("TCP server listening on %s", addrs)
            async with server:
                await server.serve_forever()
        else:
            loop = asyncio.get_running_loop()
            transport, protocol = await loop.create_datagram_endpoint(
                lambda: UDPHandler(self), local_addr=(self.host, self.port)
            )
            logging.info("UDP server listening on %s:%d", self.host, self.port)
            try:
                while True:
                    await asyncio.sleep(3600)
            finally:
                transport.close()

    async def handle_tcp(
        self, reader: asyncio.StreamReader, writer: asyncio.StreamWriter
    ):
        addr = writer.get_extra_info("peername")
        logging.info("TCP conn from %s", addr)
        try:
            buf = bytearray()
            while True:
                chunk = await reader.read(1024)
                if not chunk:
                    break
                buf.extend(chunk)
                logging.info("Received (%s): %r", addr, chunk)
                while True:
                    if len(buf) >= 1 and buf[0] == START:
                        if len(buf) < 6:
                            break
                        L1 = buf[1]
                        L2 = buf[2]
                        if L1 != L2:
                            logging.warning(
                                "Length bytes mismatch (%d != %d), resync", L1, L2
                            )
                            buf.pop(0)
                            continue
                        L = L1
                        expected = 4 + L + 2
                        if len(buf) < expected:
                            break
                        frame = bytes(buf[:expected])
                        try:
                            info = parse_ft12_frame(frame)
                            logging.info(
                                "Parsed FT1.2 request from %s: control=0x%02X addr=0x%02X payload=%r",
                                addr,
                                info["control"],
                                info["address"],
                                info["payload"],
                            )
                            resp_payload = make_ft12_time_payload()
                            resp = build_ft12_frame(
                                resp_payload, control=0x73, address=info["address"]
                            )
                        except Exception as e:
                            logging.warning(
                                "FT1.2 parse failed: %s; fallback to text", e
                            )
                            resp = make_text_response()
                        writer.write(resp)
                        await writer.drain()
                        logging.info("Sent (%s): %r", addr, resp)
                        buf = buf[expected:]
                        continue
                    else:
                        if b"\n" in buf:
                            idx = buf.find(b"\n")
                            line = bytes(buf[: idx + 1])
                            try:
                                s = line.decode("utf-8", errors="ignore").strip()
                            except:
                                s = ""
                            logging.info("Text line from %s: %r", addr, s)
                            if s.upper() == "GETTIME":
                                resp = make_text_response()
                            else:
                                resp = make_text_response()
                            writer.write(resp)
                            await writer.drain()
                            logging.info("Sent (%s): %r", addr, resp)
                            buf = buf[idx + 1 :]
                            continue
                        else:
                            break
        except Exception:
            logging.exception("TCP handler error")
        finally:
            logging.info("Connection closed %s", addr)
            try:
                writer.close()
                await writer.wait_closed()
            except:
                pass


def parse_args():
    p = argparse.ArgumentParser(
        description="Simple Teleport K-104 emulator (FT1.2-like)"
    )
    p.add_argument("--host", default="0.0.0.0")
    p.add_argument("--port", type=int, default=9000)
    p.add_argument("--proto", choices=["tcp", "udp"], default="tcp")
    p.add_argument("--mode", choices=["text", "binary"], default="text")
    return p.parse_args()


if __name__ == "__main__":
    args = parse_args()
    em = Emulator(args.host, args.port, args.proto, args.mode)
    try:
        asyncio.run(em.start())
    except KeyboardInterrupt:
        logging.info("Stopped by user")
