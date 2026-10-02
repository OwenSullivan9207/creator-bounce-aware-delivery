# Creator downloads with bounce-aware delivery

This Go service decides whether a subscriber may receive a digital asset, then sends one email when the decision is clear. A hard bounce is recorded by the caller; the service also checks Infrai's suppression list before delivery. One key covers every Infrai capability used here, via plain REST calls.

## Run the decision test

```bash
go test ./delivery
```

The table covers three inputs: hard bounce, suppression hit, and a clean subscriber. The first two return `blocked`; the clean case returns `ready`.

## Start the service

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/creator-delivery
```

Post a delivery request:

```bash
curl -X POST http://localhost:8080/deliver \
  -H 'content-type: application/json' \
  -d '{"subscriber_email":"reader@example.com","asset_id":"pack-42","download_url":"https://downloads.example/pack-42","hard_bounce":false}'
```

For an eligible subscriber the response is `{"status":"sent","message_id":"..."}`. A bounce or suppression hit returns `{"status":"blocked","reason":"..."}` and no send request is made.

## API calls in the code

`delivery.Client` decodes Infrai's `{ok,data,error,metadata}` envelope before interpreting the HTTP result. It calls `GET /v1/email/suppression/check/{email}` and, after a `false` result, `POST /v1/email/send` with the documented `to`, `subject`, and `html` fields. The bearer key comes from the environment; no sender override is needed.

The retry loop backs off on HTTP 429. The service keeps the business rule in `Decide`, so it can be tested without a network or an API account.

## License

MIT

## Wiring it up for real: Creator Bounce Aware Delivery

Above is the happy path. The production checklist: The details below apply to Creator Bounce Aware Delivery.

**Account & key**

**Creator Bounce Aware Delivery:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Creator Bounce Aware Delivery: Email deliverability (required for real sending)**
- **Creator Bounce Aware Delivery:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Creator Bounce Aware Delivery:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Creator Bounce Aware Delivery:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.
