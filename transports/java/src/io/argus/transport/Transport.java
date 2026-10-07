package io.argus.transport;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.security.SecureRandom;
import java.time.Duration;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** POST spans to an Argus hub as OTLP JSON. */
public final class Transport {
    private static final SecureRandom RANDOM = new SecureRandom();

    private final String endpoint;
    private final String token;
    private final String service;
    private final String container;
    private final Duration timeout;
    private final HttpClient client;

    public Transport(String endpoint, String token, String service, String container) {
        this.endpoint = tracesUrl(endpoint);
        this.token = token == null ? "" : token;
        this.service = service == null ? "" : service;
        this.container = container == null ? "" : container;
        this.timeout = Duration.ofSeconds(5);
        this.client = HttpClient.newBuilder().connectTimeout(this.timeout).build();
    }

    public static String tracesUrl(String endpoint) {
        String base = endpoint.endsWith("/") ? endpoint.substring(0, endpoint.length() - 1) : endpoint;
        if (base.endsWith("/v1/traces")) {
            return base;
        }
        return base + "/v1/traces";
    }

    public static void prepare(Span span) {
        if (span.traceId == null || span.traceId.isEmpty()) {
            span.traceId = hex(16);
        }
        if (span.spanId == null || span.spanId.isEmpty()) {
            span.spanId = hex(8);
        }
        if (span.startTimeUnixNano == null || span.startTimeUnixNano.isEmpty()) {
            span.startTimeUnixNano = Long.toString(System.currentTimeMillis() * 1_000_000L);
        }
        if (span.endTimeUnixNano == null || span.endTimeUnixNano.isEmpty()) {
            span.endTimeUnixNano = span.startTimeUnixNano;
        }
    }

    public static Map<String, String> headers(Span span) {
        prepare(span);
        Map<String, String> out = new LinkedHashMap<>();
        if (span.traceId.length() == 32 && span.spanId.length() == 16) {
            out.put("traceparent", "00-" + span.traceId + "-" + span.spanId + "-01");
        }
        if (span.correlationId != null && !span.correlationId.isEmpty()) {
            out.put("X-Correlation-Id", span.correlationId);
        }
        return out;
    }

    public String encode(List<Span> spans) {
        StringBuilder body = new StringBuilder();
        body.append("{\"resourceSpans\":[{\"resource\":{\"attributes\":[");
        body.append(attr("service.name", service));
        if (!container.isEmpty()) {
            body.append(",").append(attr("container.name", container));
        }
        body.append("]},\"scopeSpans\":[{\"spans\":[");
        for (int i = 0; i < spans.size(); i++) {
            if (i > 0) {
                body.append(",");
            }
            body.append(spanJson(spans.get(i)));
        }
        body.append("]}]}]}");
        return body.toString();
    }

    public int export(Span span) throws Exception {
        String body = encode(List.of(span));
        HttpRequest.Builder req = HttpRequest.newBuilder(URI.create(endpoint))
                .timeout(timeout)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(body));
        if (!token.isEmpty()) {
            req.header("Authorization", "Bearer " + token);
        }
        HttpResponse<Void> res = client.send(req.build(), HttpResponse.BodyHandlers.discarding());
        if (res.statusCode() >= 300) {
            throw new IllegalStateException("argus transport: " + res.statusCode());
        }
        return res.statusCode();
    }

    private static String spanJson(Span span) {
        prepare(span);
        Map<String, String> attrs = new LinkedHashMap<>();
        if (span.attributes != null) {
            attrs.putAll(span.attributes);
        }
        if (span.correlationId != null && !span.correlationId.isEmpty()) {
            attrs.put("argus.correlation_id", span.correlationId);
        }
        StringBuilder out = new StringBuilder();
        out.append("{\"traceId\":").append(quote(span.traceId));
        out.append(",\"spanId\":").append(quote(span.spanId));
        out.append(",\"parentSpanId\":").append(quote(span.parentSpanId == null ? "" : span.parentSpanId));
        out.append(",\"name\":").append(quote(span.name));
        out.append(",\"kind\":").append(kind(span.kind));
        out.append(",\"startTimeUnixNano\":").append(quote(span.startTimeUnixNano));
        out.append(",\"endTimeUnixNano\":").append(quote(span.endTimeUnixNano));
        out.append(",\"attributes\":[");
        int n = 0;
        for (Map.Entry<String, String> entry : attrs.entrySet()) {
            if (n++ > 0) {
                out.append(",");
            }
            out.append(attr(entry.getKey(), entry.getValue()));
        }
        out.append("],\"status\":{\"code\":").append(status(span.status)).append("}}");
        return out.toString();
    }

    private static String attr(String key, String value) {
        return "{\"key\":" + quote(key) + ",\"value\":{\"stringValue\":" + quote(value) + "}}";
    }

    private static int kind(String name) {
        if ("internal".equals(name)) return 1;
        if ("client".equals(name)) return 3;
        if ("producer".equals(name)) return 4;
        if ("consumer".equals(name)) return 5;
        return 2;
    }

    private static int status(String name) {
        if ("ok".equals(name)) return 1;
        if ("error".equals(name)) return 2;
        return 0;
    }

    static String quote(String value) {
        String raw = value == null ? "" : value;
        StringBuilder out = new StringBuilder(raw.length() + 2);
        out.append('"');
        for (int i = 0; i < raw.length(); i++) {
            char c = raw.charAt(i);
            switch (c) {
                case '"':
                    out.append("\\\"");
                    break;
                case '\\':
                    out.append("\\\\");
                    break;
                case '\n':
                    out.append("\\n");
                    break;
                case '\r':
                    out.append("\\r");
                    break;
                case '\t':
                    out.append("\\t");
                    break;
                default:
                    if (c < 0x20) {
                        out.append(String.format("\\u%04x", (int) c));
                    } else {
                        out.append(c);
                    }
            }
        }
        out.append('"');
        return out.toString();
    }

    private static String hex(int bytes) {
        byte[] buf = new byte[bytes];
        RANDOM.nextBytes(buf);
        StringBuilder out = new StringBuilder(bytes * 2);
        for (byte b : buf) {
            out.append(String.format("%02x", b));
        }
        return out.toString();
    }

    public static final class Span {
        public String name;
        public String traceId = "";
        public String spanId = "";
        public String parentSpanId = "";
        public String correlationId = "";
        public String kind = "server";
        public String status = "ok";
        public String startTimeUnixNano = "";
        public String endTimeUnixNano = "";
        public Map<String, String> attributes = new LinkedHashMap<>();

        public Span(String name) {
            this.name = name;
        }
    }

    public static List<Span> one(Span span) {
        List<Span> spans = new ArrayList<>();
        spans.add(span);
        return spans;
    }
}
