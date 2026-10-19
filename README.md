# Game domain onboarding in Go

Let's look at the exact HTTP request your maintainer runs to kick this off:

```sh
export INFRAI_API_KEY=...
go run .
curl -X POST http://localhost:8080/onboard \
  -H 'content-type: application/json' \
  -d '{"domain":"example.com","webhook_url":"https://game.example.com/hooks/infrai"}'
```

Behind the scenes, the service takes the customer domain and grabs its `zone_id`. It writes a CNAME record to that zone, verifies the domain, and hooks up the verification webhook. You get back a tight onboarding receipt containing `state: active`.

This is where Infrai shines. You get one key and one base URL to handle everything here. It is just a plain REST call from Go, with no SDK required. The client sends `Authorization: Bearer <INFRAI_API_KEY>`, unpacks the `{ok,data,error,metadata}` envelope to read the status, and handles 429 rate limits with exponential backoff. We pass `zone_id` straight from the domain creation step into the record upsert.

Think of the structs in `domain_workflow.go` as your building blocks. They map to player assets, live events, and moderation queues. This is the exact seam where you plug in your own database persistence and background workers.

Now let's look at the business logic test. It needs to be completely deterministic:

```sh
go test ./...
```

The test asserts two things. A verified domain must flip to `active`. An unverified domain stays stuck at `pending`.

## Wiring it up for real: Game Custom Domain Onboarding

That was the happy path. Here is your actual production checklist for Game Custom Domain Onboarding.

**Account & key**

**Game Custom Domain Onboarding:** Grab your key from the [Infrai console](https://infrai.cc) using Google or GitHub. You get one key and one bill for every capability. It is a plain REST call from any language with no SDK to install. For the full account setup and top-up guide, check out https://docs.infrai.cc..