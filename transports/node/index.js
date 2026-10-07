import { randomBytes } from "node:crypto"

const KINDS = { internal: 1, server: 2, client: 3, producer: 4, consumer: 5 }
const STATUS = { unset: 0, ok: 1, error: 2 }

export function tracesUrl(endpoint) {
  const base = endpoint.replace(/\/$/, "")
  return base.endsWith("/v1/traces") ? base : `${base}/v1/traces`
}

export function nowNano() {
  return (BigInt(Date.now()) * 1000000n).toString()
}

export function prepare(span) {
  if (!span.traceId) span.traceId = randomBytes(16).toString("hex")
  if (!span.spanId) span.spanId = randomBytes(8).toString("hex")
  if (!span.startTimeUnixNano) span.startTimeUnixNano = nowNano()
  if (!span.endTimeUnixNano) span.endTimeUnixNano = span.startTimeUnixNano
  return span
}

export function headers(span) {
  prepare(span)
  const out = {}
  if (span.traceId.length === 32 && span.spanId.length === 16) {
    out.traceparent = `00-${span.traceId}-${span.spanId}-01`
  }
  if (span.correlationId) out["X-Correlation-Id"] = span.correlationId
  return out
}

export function encode(transport, spans) {
  const encoded = spans.map((span) => {
    prepare(span)
    const attrs = { ...(span.attributes ?? {}) }
    if (span.correlationId) attrs["argus.correlation_id"] = span.correlationId
    return {
      traceId: span.traceId,
      spanId: span.spanId,
      parentSpanId: span.parentSpanId ?? "",
      name: span.name,
      kind: KINDS[span.kind] ?? KINDS.server,
      startTimeUnixNano: String(span.startTimeUnixNano),
      endTimeUnixNano: String(span.endTimeUnixNano),
      attributes: Object.entries(attrs).map(([key, value]) => ({
        key,
        value: { stringValue: String(value) },
      })),
      status: { code: STATUS[span.status] ?? STATUS.unset },
    }
  })
  const resource = [{ key: "service.name", value: { stringValue: transport.service ?? "" } }]
  if (transport.container) {
    resource.push({ key: "container.name", value: { stringValue: transport.container } })
  }
  return { resourceSpans: [{ resource: { attributes: resource }, scopeSpans: [{ spans: encoded }] }] }
}

export function createTransport({ endpoint, token = "", service = "", container = "", timeoutMs = 5000, fetchImpl = fetch }) {
  const url = tracesUrl(endpoint)
  return {
    endpoint: url,
    token,
    service,
    container,
    headers,
    encode(spans) {
      return encode(this, spans)
    },
    async export(span) {
      const body = JSON.stringify(encode(this, [span]))
      const reqHeaders = { "Content-Type": "application/json" }
      if (token) reqHeaders.Authorization = `Bearer ${token}`
      const res = await fetchImpl(url, {
        method: "POST",
        headers: reqHeaders,
        body,
        signal: AbortSignal.timeout(timeoutMs),
      })
      if (!res.ok) throw new Error(`argus transport: ${res.status}`)
      return res.status
    },
  }
}
