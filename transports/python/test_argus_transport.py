import json
import unittest

from argus_transport import Span, Transport


class EncodeTest(unittest.TestCase):
    def test_otlp_json_shape(self):
        transport = Transport("http://hub:8080", token="secret", service="orders", container="orders-1")
        span = Span(
            name="GET /orders",
            trace_id="5b8efff798038103d269b633813fc60c",
            span_id="7a1b2c3d4e5f6071",
            parent_span_id="0011223344556677",
            correlation_id="corr-1",
            kind="client",
            status="error",
            start_time_unix_nano="1000",
            end_time_unix_nano="2000",
            attributes={"http.route": "/orders"},
        )
        body = transport.encode([span])
        raw = json.dumps(body)
        self.assertIn("/v1/traces", transport.endpoint)
        span_json = body["resourceSpans"][0]["scopeSpans"][0]["spans"][0]
        self.assertEqual(span_json["kind"], 3)
        self.assertEqual(span_json["status"]["code"], 2)
        self.assertEqual(span_json["startTimeUnixNano"], "1000")
        self.assertEqual(span_json["parentSpanId"], "0011223344556677")
        keys = {item["key"]: item["value"]["stringValue"] for item in span_json["attributes"]}
        self.assertEqual(keys["argus.correlation_id"], "corr-1")
        self.assertEqual(keys["http.route"], "/orders")
        resource = body["resourceSpans"][0]["resource"]["attributes"]
        self.assertEqual(resource[0]["value"]["stringValue"], "orders")
        self.assertIn("corr-1", raw)
        headers = transport.headers(span)
        self.assertEqual(headers["traceparent"], "00-5b8efff798038103d269b633813fc60c-7a1b2c3d4e5f6071-01")
        self.assertEqual(headers["X-Correlation-Id"], "corr-1")

    def test_fills_ids(self):
        transport = Transport("http://hub:8080/v1/traces", service="orders")
        span = Span(name="GET /orders")
        transport.prepare(span)
        self.assertEqual(len(span.trace_id), 32)
        self.assertEqual(len(span.span_id), 16)
        self.assertTrue(span.start_time_unix_nano.isdigit())


if __name__ == "__main__":
    unittest.main()
