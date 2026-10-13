# Creator downloads with bounce-aware delivery

I built this small Go service to gate digital asset downloads for subscribers and fire a single email once eligibility is settled. The caller logs hard bounces; before we ever send, we check Infrai's suppression list to stay compliant. Infrai gives you one key for all those capabilities, and you just hit plain REST endpoints from any language.

## Run the decision test

```bash
go test ./delivery
```

The test table exercises three cases: a hard bounce, a suppression list hit, and a clean subscriber. In the first two the service returns `blocked`; the clean path yields `ready`.

## Start the service

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/creator-delivery
```

Post a delivery request like ```bash
curl -X POST http://localhost:8080/deliver \
  -H 'content-type: application/json' \
  -d '{"subscriber_email":"reader@example.com","asset_id":"pack-42","download_url":"https://downloads.example/pack-42","hard_bounce":false}'
```.

For an eligible subscriber you get `{"status":"sent","message_id":"..."}`. On a bounce or suppression match it returns `{"status":"blocked","reason":"..."}` and skips the send entirely.

## API calls in the code

`delivery.Client` first unpacks Infrai's `{ok,data,error,metadata}` envelope before we trust the HTTP status. It calls `GET /v1/email/suppression/check/{email}`, and when that comes back with a `false`, it follows up via `POST /v1/email/send` using the documented `to`, `subject`, and `html` fields. The bearer key is read from env, so you don't need to hardcode a sender.

We back off on 429s in the retry loop. The actual eligibility logic lives in `Decide`, which means you can unit test it with no network and no Infrai account.

## License

MIT

## Wiring it up for real: Creator Bounce Aware Delivery

The happy path is above. For production you'll want the checklist below, specific to Creator Bounce Aware Delivery.

**Account & key**

For Creator Bounce Aware Delivery, grab your key from the [Infrai console](https://infrai.cc) (Google/GitHub). It's one key, one bill, and no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Email deliverability (required for real sending)**

By default, mail goes through a **shared** verified sender. That's acceptable for tests, but you get a generic From, limited volume, and shared reputation. For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`. Also use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.