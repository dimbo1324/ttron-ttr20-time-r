import argparse
import socket
import time
import sys
import logging
from datetime import datetime

START = 0x68
END = 0x16


def build_ft12_frame(payload: bytes, control: int = 0x53, address: int = 0x01) -> bytes:
    L = 1 + 1 + len(payload)
    f = bytearray()
    f.append(START)
    f.append(L & 0xFF)
    f.append(L & 0xFF)
    f.append(START)
    f.append(control & 0xFF)
    f.append(address & 0xFF)
    f.extend(payload)
    cs = sum(f[4:]) & 0xFF
    f.append(cs)
    f.append(END)
    return bytes(f)


def parse_ft12_frame(buf: bytes):
    if len(buf) < 6:
        raise ValueError("too short")
    if buf[0] != START or buf[3] != START:
        raise ValueError("bad start")
    L1 = buf[1]
    L2 = buf[2]
    if L1 != L2:
        raise ValueError("length mismatch")
    L = L1
    expected = 4 + L + 2
    if len(buf) != expected:
        raise ValueError("len mismatch")
    control = buf[4]
    address = buf[5]
    payload = buf[6 : 6 + (L - 2)]
    cs = buf[6 + (L - 2)]
    calc = sum(buf[4 : 6 + (L - 2)]) & 0xFF
    if cs != calc:
        raise ValueError("checksum mismatch")
    if buf[-1] != END:
        raise ValueError("end mismatch")
    return {"control": control, "address": address, "payload": bytes(payload)}


def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument("--host", default="127.0.0.1")
    p.add_argument("--port", type=int, default=9000)
    p.add_argument("--proto", choices=["tcp", "udp"], default="tcp")
    p.add_argument("--log", default="device_time.log")
    p.add_argument("--timeout", type=float, default=2.0)
    p.add_argument("--retry", type=float, default=3.0)
    return p.parse_args()


def setup_logging(logfile):
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(message)s",
        handlers=[
            logging.FileHandler(logfile, encoding="utf-8"),
            logging.StreamHandler(sys.stdout),
        ],
    )


def seconds_until_next_5s():
    now = time.time()
    sec = int(now) % 60
    offset = (5 - (sec % 5)) % 5
    target = int(now) + offset
    frac = now - int(now)
    if offset == 0 and frac > 0.15:
        target += 5
    return max(0.0, target - now)


def recv_exact(sock, n, timeout):
    sock.settimeout(timeout)
    data = bytearray()
    try:
        while len(data) < n:
            chunk = sock.recv(n - len(data))
            if not chunk:
                raise ConnectionError("closed")
            data.extend(chunk)
        return bytes(data)
    finally:
        sock.settimeout(None)


def recv_ft12_from_tcp(sock, timeout):
    hdr = recv_exact(sock, 4, timeout)
    if hdr[0] != START or hdr[3] != START:
        rest = b""
        try:
            while True:
                ch = recv_exact(sock, 1, 0.2)
                rest += ch
                if ch == b"\n":
                    break
        except:
            pass
        return ("text", (hdr + rest).decode(errors="replace"))
    L = hdr[1]
    tail = recv_exact(sock, L + 2, timeout)
    full = hdr + tail
    info = parse_ft12_frame(full)
    return ("ft12", info)


def main():
    args = parse_args()
    setup_logging(args.log)
    logger = logging.getLogger()
    addr = (args.host, args.port)
    logger.info(
        "Client starting; target %s:%d proto=%s", args.host, args.port, args.proto
    )

    tcp_sock = None
    udp_sock = None

    try:
        while True:
            if args.proto == "tcp" and tcp_sock is None:
                try:
                    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                    s.settimeout(5.0)
                    s.connect(addr)
                    s.settimeout(None)
                    tcp_sock = s
                    logger.info("Connected TCP to %s:%d", args.host, args.port)
                except Exception as e:
                    logger.warning(
                        "connect failed: %s ; retrying in %.1fs", e, args.retry
                    )
                    time.sleep(args.retry)
                    continue
            if args.proto == "udp" and udp_sock is None:
                udp_sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
                logger.info("UDP socket ready")

            wait = seconds_until_next_5s()
            time.sleep(wait)

            try:
                if args.proto == "tcp":
                    payload = b"REQTIME"
                    frame = build_ft12_frame(payload, control=0x53, address=0x01)
                    tcp_sock.sendall(frame)
                    kind, resp = recv_ft12_from_tcp(tcp_sock, args.timeout)
                    if kind == "ft12":
                        payload = resp["payload"]
                        out = payload.decode(errors="replace")
                    else:
                        out = resp.strip()
                else:
                    payload = b"REQTIME"
                    frame = build_ft12_frame(payload, control=0x53, address=0x01)
                    udp_sock.sendto(frame, addr)
                    udp_sock.settimeout(args.timeout)
                    try:
                        data, _ = udp_sock.recvfrom(8192)
                    except socket.timeout:
                        out = "<timeout>"
                    else:
                        try:
                            info = parse_ft12_frame(data)
                            out = info["payload"].decode(errors="replace")
                        except Exception:
                            try:
                                out = data.decode(errors="replace")
                            except:
                                out = repr(data)
                    udp_sock.settimeout(None)

                ts = datetime.now().isoformat(timespec="seconds")
                logger.info("%s | %s", ts, out)
            except Exception as e:
                logger.warning("Comm error: %s", e)
                if tcp_sock:
                    try:
                        tcp_sock.close()
                    except:
                        pass
                tcp_sock = None
                time.sleep(args.retry)
                continue

            time.sleep(0.05)

    except KeyboardInterrupt:
        logger.info("Stopped by user")
    finally:
        try:
            if tcp_sock:
                tcp_sock.close()
            if udp_sock:
                udp_sock.close()
        except:
            pass


if __name__ == "__main__":
    main()
