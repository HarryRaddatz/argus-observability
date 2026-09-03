#!/usr/bin/env python3
"""Docker API unix proxy: only containers named venuz-*."""

from __future__ import annotations

import http.client
import http.server
import json
import os
import socket
import socketserver
import sys
import threading
import urllib.parse

DOCKER_SOCK = os.environ.get("DOCKER_SOCK", "/var/run/docker.sock")
LISTEN_SOCK = os.environ.get("LISTEN_SOCK", "/beszel_socket/docker-venuz.sock")
PREFIX = os.environ.get("NAME_PREFIX", "venuz-")

_allowed_lock = threading.Lock()
_allowed_ids: set[str] = set()


class UnixHTTPConnection(http.client.HTTPConnection):
    def __init__(self, sock_path: str, timeout: float | None = 30.0) -> None:
        super().__init__("localhost", timeout=timeout)
        self._sock_path = sock_path

    def connect(self) -> None:
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        if self.timeout is not None:
            sock.settimeout(self.timeout)
        sock.connect(self._sock_path)
        self.sock = sock


def docker_name_allowed(name: str) -> bool:
    return name.lstrip("/").startswith(PREFIX)


def remember_container(entry: dict) -> None:
    cid = str(entry.get("Id") or "")
    with _allowed_lock:
        if cid:
            _allowed_ids.add(cid)
            _allowed_ids.add(cid[:12])
        for name in entry.get("Names") or []:
            stripped = str(name).lstrip("/")
            if stripped:
                _allowed_ids.add(stripped)
                _allowed_ids.add(name)


def id_allowed(container_id: str) -> bool:
    with _allowed_lock:
        if container_id in _allowed_ids:
            return True
    conn = UnixHTTPConnection(DOCKER_SOCK)
    try:
        conn.request("GET", f"/containers/{urllib.parse.quote(container_id, safe='')}/json")
        resp = conn.getresponse()
        body = resp.read()
        if resp.status != 200:
            return False
        info = json.loads(body.decode("utf-8"))
        name = str(info.get("Name") or "")
        names = info.get("Name")
        allowed = docker_name_allowed(name)
        if allowed:
            remember_container({"Id": info.get("Id", container_id), "Names": [name] if name else []})
        return allowed
    except Exception:
        return False
    finally:
        conn.close()


def filter_container_list(body: bytes) -> bytes:
    data = json.loads(body.decode("utf-8"))
    if not isinstance(data, list):
        return body
    kept = []
    with _allowed_lock:
        _allowed_ids.clear()
    for entry in data:
        names = entry.get("Names") or []
        if any(docker_name_allowed(str(n)) for n in names):
            kept.append(entry)
            remember_container(entry)
    return json.dumps(kept).encode("utf-8")


def path_container_id(path: str) -> str | None:
    parsed = urllib.parse.urlparse(path)
    parts = parsed.path.split("/")
    try:
        idx = parts.index("containers")
    except ValueError:
        return None
    if idx + 1 >= len(parts):
        return None
    ident = parts[idx + 1]
    if ident in ("json", "create"):
        return None
    return urllib.parse.unquote(ident)


def is_container_list(path: str) -> bool:
    parsed = urllib.parse.urlparse(path)
    trimmed = parsed.path.rstrip("/")
    return trimmed.endswith("/containers/json") or trimmed == "/containers/json"


def is_streaming_path(path: str) -> bool:
    parsed = urllib.parse.urlparse(path)
    trimmed = parsed.path.rstrip("/")
    if trimmed.endswith("/events") or trimmed == "/events":
        return True
    if "/logs" in trimmed:
        qs = urllib.parse.parse_qs(parsed.query)
        follow = (qs.get("follow") or ["false"])[0].lower()
        if follow in ("1", "true", "yes"):
            return True
    return False


class Handler(http.server.BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def log_message(self, fmt: str, *args) -> None:
        sys.stderr.write("%s\n" % (fmt % args))

    def do_GET(self) -> None:
        self._forward("GET")

    def do_HEAD(self) -> None:
        self._forward("HEAD")

    def _forward(self, method: str) -> None:
        cid = path_container_id(self.path)
        is_list = is_container_list(self.path)

        if cid and not is_list and not id_allowed(cid):
            self.send_error(404, "Not Found")
            return

        if is_streaming_path(self.path):
            self._forward_stream(method)
            return

        headers = {k: v for k, v in self.headers.items() if k.lower() not in {"host", "connection"}}
        conn = UnixHTTPConnection(DOCKER_SOCK)
        try:
            conn.request(method, self.path, headers=headers)
            resp = conn.getresponse()
            body = resp.read()
        except Exception as exc:
            self.send_error(502, str(exc))
            return
        finally:
            conn.close()

        if is_list and resp.status == 200 and method == "GET":
            try:
                body = filter_container_list(body)
            except Exception as exc:
                self.send_error(502, str(exc))
                return

        self.send_response(resp.status, resp.reason)
        skip = {"transfer-encoding", "connection", "content-length"}
        for key, value in resp.getheaders():
            if key.lower() not in skip:
                self.send_header(key, value)
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Connection", "close")
        self.end_headers()
        if method != "HEAD":
            self.wfile.write(body)

    def _forward_stream(self, method: str) -> None:
        headers = {k: v for k, v in self.headers.items() if k.lower() not in {"host", "connection"}}
        conn = UnixHTTPConnection(DOCKER_SOCK, timeout=None)
        try:
            conn.request(method, self.path, headers=headers)
            resp = conn.getresponse()
            self.send_response(resp.status, resp.reason)
            skip = {"transfer-encoding", "connection", "content-length"}
            for key, value in resp.getheaders():
                if key.lower() not in skip:
                    self.send_header(key, value)
            self.send_header("Connection", "close")
            self.end_headers()
            if method != "HEAD":
                while True:
                    chunk = resp.read(8192)
                    if not chunk:
                        break
                    self.wfile.write(chunk)
                    self.wfile.flush()
        except Exception as exc:
            self.send_error(502, str(exc))
        finally:
            conn.close()


class UnixHTTPServer(socketserver.ThreadingMixIn, socketserver.UnixStreamServer):
    daemon_threads = True
    allow_reuse_address = True


def main() -> int:
    if os.path.exists(LISTEN_SOCK):
        os.remove(LISTEN_SOCK)
    os.makedirs(os.path.dirname(LISTEN_SOCK), exist_ok=True)
    server = UnixHTTPServer(LISTEN_SOCK, Handler)
    os.chmod(LISTEN_SOCK, 0o660)
    sys.stderr.write(f"listening {LISTEN_SOCK} prefix={PREFIX}\n")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()
        if os.path.exists(LISTEN_SOCK):
            os.remove(LISTEN_SOCK)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
