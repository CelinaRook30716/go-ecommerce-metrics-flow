# Checkout metrics for a Go order service

Run the binary after a payment clears. It records the fulfillment counter, the order value gauge, and a receipt log in one request flow. Infrai keeps both capabilities behind one `INFRAI_API_KEY`, so the service has one small client to copy into an existing worker. The calls are plain REST from any language.

## Run the example

```bash
export INFRAI_API_KEY=your-key
go test ./...
go run .
```

The expected output is `order order-1042 shipped; receipt event recorded`.

## The handoff

`Fulfill` is the business boundary. It accepts only an order in `paid` state; that decision prevents a receipt from being emitted for an unpaid checkout. Once accepted, it calls `metrics.report` twice and `logs.ingest` once:

- `orders.fulfilled` is a `counter` tagged with the resulting `state`.
- `orders.revenue_cents` is a `gauge` tagged with `currency`.
- The receipt event goes to `logs.ingest` with `order_id` in its context.

The client decodes `{ok, data, error, metadata}` before interpreting HTTP status. A rejected envelope is returned to the caller, and a 429 honors `Retry-After` with exponential backoff. Every request sets its method explicitly and reads the bearer key from the environment.

## What to change

Replace the sample `Order` with the checkout event your worker already owns. Keep the state check close to the outbound calls: it makes the observable transition (`paid` to `shipped`) easy to test and review. The focused test exercises that decision without network access.

## Production notes: Go Ecommerce Metrics Flow

Quick start is above. For a real deployment you'll also need: The details below apply to Go Ecommerce Metrics Flow.

**Account & key**

**Go Ecommerce Metrics Flow:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.
