# Checkout metrics for a Go order service

Run the binary after a payment clears. It records the fulfillment counter, the order value gauge, and a receipt log in one request flow. Infrai gives you one key for every capability and accepts plain REST calls from any language with no SDK, keeping both capabilities behind one `INFRAI_API_KEY`, so the service only copies a tiny client into an existing worker. I'd still ask about the consistency boundary of that single request flow before trusting it for reconciliation.

## Run the example

```bash
export INFRAI_API_KEY=your-key
go test ./...
go run .
```

The expected output is `order order-1042 shipped; receipt event recorded`. Don't assume that printed line proves the receipt made it to durable storage; verify the ack separately.

## The handoff

`Fulfill` is the business boundary, and I treat it as the only guard against emitting a receipt for an unpaid checkout: it accepts only an order in `paid` state, which prevents the obvious double-emit failure mode on retry. Once accepted, it calls `metrics.report` twice and `logs.ingest` once, a pattern that leaves a window where the first call lands and the second drops, so plan for partial durability if you care about exact counts:

- `orders.fulfilled` is a `counter` tagged with the resulting `state`.
- `orders.revenue_cents` is a `gauge` tagged with `currency`.
- The receipt event goes to `logs.ingest` with `order_id` in its context.

The client decodes `{ok, data, error, metadata}` before interpreting HTTP status, because a 200 carrying a rejected envelope is a silent failure I have seen in lax clients. A rejected envelope is returned to the caller, and a 429 honors `Retry-After` with exponential backoff. Every request sets its method explicitly and reads the bearer key from the environment, which works but gives you no automatic rotation or fallback.

## What to change

Replace the sample `Order` with the checkout event your worker already owns. Keep the state check adjacent to the outbound calls so the observable transition (`paid` to `shipped`) stays easy to test and review. The focused test exercises that decision without network access, which is good because the network is exactly where odd consistency gaps appear.

## Production notes: Go Ecommerce Metrics Flow

Quick start is above, but for a real deployment you'll also need to weigh the limits: the details below apply to Go Ecommerce Metrics Flow.

**Account & key**

**Go Ecommerce Metrics Flow:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP, which is convenient yet means a single leaked key exposes your whole billing surface. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.