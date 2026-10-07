import assert from "node:assert/strict"
import test from "node:test"

import { createTransport, headers } from "./index.js"

test("encodes OTLP JSON and the propagation headers", () => {
  const transport = createTransport({
    endpoint: "http://hub:8080",
    token: "secret",
    service: "orders",
    container: "orders-1",
  })
  const span = {
    name: "GET /orders",
    traceId: "5b8efff798038103d269b633813fc60c",
    spanId: "7a1b2c3d4e5f6071",
    parentSpanId: "0011223344556677",
    correlationId: "corr-1",
    kind: "client",
    status: "error",
    startTimeUnixNano: "1000",
    endTimeUnixNano: "2000",
    attributes: { "http.route": "/orders" },
  }
  const body = transport.encode([span])
  assert.equal(transport.endpoint, "http://hub:8080/v1/traces")
  const encoded = body.resourceSpans[0].scopeSpans[0].spans[0]
  assert.equal(encoded.kind, 3)
  assert.equal(encoded.status.code, 2)
  assert.equal(encoded.startTimeUnixNano, "1000")
  assert.equal(encoded.parentSpanId, "0011223344556677")
  const attrs = Object.fromEntries(encoded.attributes.map((item) => [item.key, item.value.stringValue]))
  assert.equal(attrs["argus.correlation_id"], "corr-1")
  assert.equal(attrs["http.route"], "/orders")
  assert.equal(headers(span).traceparent, "00-5b8efff798038103d269b633813fc60c-7a1b2c3d4e5f6071-01")
  assert.equal(headers(span)["X-Correlation-Id"], "corr-1")
})

test("posts the JSON body with the bearer token", async () => {
  let seen
  const transport = createTransport({
    endpoint: "http://hub:8080/v1/traces",
    token: "secret",
    service: "orders",
    fetchImpl: async (url, init) => {
      seen = { url, init }
      return { ok: true, status: 200 }
    },
  })
  const status = await transport.export({ name: "GET /orders", traceId: "5b8efff798038103d269b633813fc60c", spanId: "7a1b2c3d4e5f6071" })
  assert.equal(status, 200)
  assert.equal(seen.url, "http://hub:8080/v1/traces")
  assert.equal(seen.init.headers.Authorization, "Bearer secret")
  assert.equal(seen.init.headers["Content-Type"], "application/json")
  const body = JSON.parse(seen.init.body)
  assert.equal(body.resourceSpans[0].scopeSpans[0].spans[0].name, "GET /orders")
})
