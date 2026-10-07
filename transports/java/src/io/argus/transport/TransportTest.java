package io.argus.transport;

public final class TransportTest {
    public static void main(String[] args) {
        Transport transport = new Transport("http://hub:8080", "secret", "orders", "orders-1");
        if (!transport.tracesUrl("http://hub:8080").equals("http://hub:8080/v1/traces")) {
            throw new AssertionError("url");
        }
        Transport.Span span = new Transport.Span("GET /orders");
        span.traceId = "5b8efff798038103d269b633813fc60c";
        span.spanId = "7a1b2c3d4e5f6071";
        span.parentSpanId = "0011223344556677";
        span.correlationId = "corr-1";
        span.kind = "client";
        span.status = "error";
        span.startTimeUnixNano = "1000";
        span.endTimeUnixNano = "2000";
        span.attributes.put("http.route", "/orders");
        String body = transport.encode(Transport.one(span));
        require(body.contains("\"kind\":3"), body);
        require(body.contains("\"code\":2"), body);
        require(body.contains("\"startTimeUnixNano\":\"1000\""), body);
        require(body.contains("\"parentSpanId\":\"0011223344556677\""), body);
        require(body.contains("\"argus.correlation_id\""), body);
        require(body.contains("\"stringValue\":\"corr-1\""), body);
        require(body.contains("\"http.route\""), body);
        require(body.contains("\"service.name\""), body);
        String parent = Transport.headers(span).get("traceparent");
        require("00-5b8efff798038103d269b633813fc60c-7a1b2c3d4e5f6071-01".equals(parent), parent);
        require("corr-1".equals(Transport.headers(span).get("X-Correlation-Id")), "correlation");
        String quoted = Transport.quote("a\"b\\c");
        require("\"a\\\"b\\\\c\"".equals(quoted), quoted);
    }

    private static void require(boolean ok, String detail) {
        if (!ok) {
            throw new AssertionError(detail);
        }
    }
}
