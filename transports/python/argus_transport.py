"""POST spans to an Argus hub as OTLP JSON."""

import json
import secrets
import time
import urllib.error
import urllib.request


def traces_url(endpoint):
    base = endpoint.rstrip("/")
    if base.endswith("/v1/traces"):
        return base
    return base + "/v1/traces"


def now_nano():
    return str(time.time_ns())


def new_trace_id():
    return secrets.token_hex(16)


def new_span_id():
    return secrets.token_hex(8)


_KINDS = {"internal": 1, "server": 2, "client": 3, "producer": 4, "consumer": 5}
_STATUS = {"unset": 0, "ok": 1, "error": 2}


class Span:
    def __init__(
        self,
        name,
        trace_id="",
        span_id="",
        parent_span_id="",
        correlation_id="",
        kind="server",
        status="ok",
        start_time_unix_nano="",
        end_time_unix_nano="",
        attributes=None,
    ):
        self.name = name
        self.trace_id = trace_id
        self.span_id = span_id
        self.parent_span_id = parent_span_id
        self.correlation_id = correlation_id
        self.kind = kind
        self.status = status
        self.start_time_unix_nano = start_time_unix_nano
        self.end_time_unix_nano = end_time_unix_nano
        self.attributes = dict(attributes or {})


class Transport:
    def __init__(self, endpoint, token="", service="", container="", timeout=5.0):
        self.endpoint = traces_url(endpoint)
        self.token = token
        self.service = service
        self.container = container
        self.timeout = timeout

    def prepare(self, span):
        if not span.trace_id:
            span.trace_id = new_trace_id()
        if not span.span_id:
            span.span_id = new_span_id()
        if not span.start_time_unix_nano:
            span.start_time_unix_nano = now_nano()
        if not span.end_time_unix_nano:
            span.end_time_unix_nano = span.start_time_unix_nano
        return span

    def headers(self, span):
        self.prepare(span)
        out = {}
        if len(span.trace_id) == 32 and len(span.span_id) == 16:
            out["traceparent"] = f"00-{span.trace_id}-{span.span_id}-01"
        if span.correlation_id:
            out["X-Correlation-Id"] = span.correlation_id
        return out

    def encode(self, spans):
        encoded = []
        for span in spans:
            self.prepare(span)
            attrs = dict(span.attributes)
            if span.correlation_id:
                attrs["argus.correlation_id"] = span.correlation_id
            encoded.append(
                {
                    "traceId": span.trace_id,
                    "spanId": span.span_id,
                    "parentSpanId": span.parent_span_id,
                    "name": span.name,
                    "kind": _KINDS.get(span.kind, 2),
                    "startTimeUnixNano": str(span.start_time_unix_nano),
                    "endTimeUnixNano": str(span.end_time_unix_nano),
                    "attributes": [
                        {"key": key, "value": {"stringValue": str(value)}}
                        for key, value in attrs.items()
                    ],
                    "status": {"code": _STATUS.get(span.status, 0)},
                }
            )
        resource = [{"key": "service.name", "value": {"stringValue": self.service}}]
        if self.container:
            resource.append({"key": "container.name", "value": {"stringValue": self.container}})
        return {
            "resourceSpans": [
                {
                    "resource": {"attributes": resource},
                    "scopeSpans": [{"spans": encoded}],
                }
            ]
        }

    def export(self, span):
        body = json.dumps(self.encode([span])).encode()
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = "Bearer " + self.token
        req = urllib.request.Request(self.endpoint, data=body, headers=headers, method="POST")
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                return resp.status
        except urllib.error.HTTPError as err:
            raise RuntimeError(f"argus transport: {err.code}") from err
