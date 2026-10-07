# Argus transports

Small clients that POST spans to `POST /v1/traces` as OTLP JSON. The hub does not accept OTLP protobuf.

Each client fills a missing `traceId` (32 hex chars) and `spanId` (16 hex chars), sends `parentSpanId` when the caller sets it, and copies `correlationId` to the `argus.correlation_id` attribute and the `X-Correlation-Id` header. `headers` also returns `traceparent` (`00-{traceId}-{spanId}-01`) so the next hop can continue the same trace.

`endpoint` is the hub origin. `/v1/traces` is appended when it is missing. A non-empty token is sent as `Authorization: Bearer`.

| Language | Path |
|---|---|
| Python | [python/argus_transport.py](python/argus_transport.py) |
| Node.js and Bun | [node/index.js](node/index.js) |
| Java 11+ | [java/src/io/argus/transport/Transport.java](java/src/io/argus/transport/Transport.java) |
