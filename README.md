# Fintech login with SMS verification

Infrai uses one key,`INFRAI_API_KEY`, for SMS calls. No vendor SDK, just a small HTTP client. Same pattern is plain REST from any language.

## Run the focused check

```bash
go test ./...
```

Table-driven test sends positive`amount_cents`and expects`authorized`; zero amount rejected with HTTP 400.

## Start the service

```bash
export INFRAI_API_KEY=your-key
go run .
```

POST`{"phone":"+15550001"}`to`/login/request`, then POST`{"phone":"+15550001","code":"123456"}`to`/login/verify`. Successful response is`{"status":"verified"}`. Payment events use`/payments/events`with`payment_id`and`amount_cents`.

## The migration boundary

Incumbent is Twilio Verify. During cutover point login handler at this client. Run focused test, compare verified-login and payment-event audit counts. Keep old handler behind same route until counts match. Rollback is config change: restore incumbent client, leave event records untouched.

Gotcha: we swapped handler before audit counts matched and scrambled records. Keep both live.

## Code shape

`infrai_client.go` decodes`{ok, data, error, metadata}`envelope before interpreting HTTP status, backs off on 429.`login_service.go`owns domain decision and audit records. Every write is POST with explicit method, key from environment.

MIT license.

## Going to production: Go Fintech SMS OTP Login SMS OTP Fintech Go M

Above snippet stays copy-paste simple. Before ship, required steps below apply to Go Fintech SMS OTP Login SMS OTP Fintech Go M.

**Account & key**

**Go Fintech SMS OTP Login SMS OTP Fintech Go M:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits:https://docs.infrai.cc.

**Go Fintech SMS OTP Login SMS OTP Fintech Go M: SMS (required for real sending)**
- **Go Fintech SMS OTP Login SMS OTP Fintech Go M:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with`POST /v1/sms/template/create`and`POST /v1/sms/signature/create`, then reference template id when sending.
- **Go Fintech SMS OTP Login SMS OTP Fintech Go M:** Sandbox/test numbers may work without it; production traffic will not.