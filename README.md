# wenova-laos-sms-services

A small, secure Go client for the [Wenova Link SMS API](https://apimicroservices.wenova.fun) — send SMS (OTP, notifications, promotions) to Lao mobile numbers in `20XXXXXXXX` format.

- **Simple** — three methods: `SendOTP`, `SendInfo`, `SendCustom`.
- **Secure by default** — your API token stays server-side and is never logged or leaked into errors.
- **Customizable** — wallet or SMS package billing, custom sender header, custom endpoint/client.
- **Fails fast** — validates phone numbers and messages before any network call.

> **v1 scope:** token (`key`) authentication only. `scriptId` is not supported yet.

---

## Use cases

### Send an OTP

```go
wnv, err := wenovasms.New(wenovasms.Config{Key: "your-api-token"})
if err != nil {
    log.Fatal(err)
}

res, err := wnv.SendOTP(context.Background(), "2012345678", "Your OTP is: 123456")
if err != nil {
    log.Fatal(err) // message rejected or gateway failed
}
fmt.Println("sent:", res.ResultCode) // "20000" = success
```

Sender header: `WNV-OTP` — handled for you.

### Send an info/notification message

```go
_, err := wnv.SendInfo(context.Background(), "2012345678", "Your order has shipped.")
```

Sender header: `WNV-info` — handled for you.

### Send with a custom sender header

```go
_, err := wnv.SendCustom(context.Background(), "WNV-Promo", "2012345678", "50% off this weekend!")
```

You provide the registered sender ID yourself.

### Bill from your SMS package (instead of wallet)

```go
wnv, err := wenovasms.New(wenovasms.Config{
    Key:        "your-api-token",
    UsePackage: true, // true = SMS package, false = wallet (default)
})
```

---

## Installation

```bash
go get github.com/barlus-developer/wenova-laos-sms-services
```

Requires **Go 1.23+**.

---

## Quick start

```go
package main

import (
    "context"
    "log"

    "github.com/barlus-developer/wenova-laos-sms-services"
)

func main() {
    wnv, err := wenovasms.New(wenovasms.Config{Key: "your-api-token"})
    if err != nil {
        log.Fatal(err)
    }

    res, err := wnv.SendOTP(context.Background(), "2012345678", "Your OTP is: 123456")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("SMS sent (resultCode=%s)", res.ResultCode)
}
```

---

## Configuration

`wenovasms.Config` — all fields optional except `Key`.

| Field        | Type          | Default                       | Description                                        |
| ------------ | ------------- | ----------------------------- | -------------------------------------------------- |
| `Key`        | `string`      | — (required)                  | API token from Dashboard → Key                     |
| `UsePackage` | `bool`        | `false` (wallet)              | `true` = SMS package, `false` = wallet             |
| `BaseURL`    | `string`      | `https://apimicroservices.wenova.fun` | Override the API endpoint (testing/private gateways) |
| `HTTPClient` | `*http.Client`| client with 10s timeout      | Inject your own HTTP client                        |

```go
cfg := wenovasms.Config{
    Key:        "your-api-token",
    UsePackage: true,          // bill from SMS package
    HTTPClient: &http.Client{Timeout: 5 * time.Second},
}
wnv, err := wenovasms.New(cfg)
```

### Billing

| `UsePackage` | Payment source                      |
| ------------ | ----------------------------------- |
| `true`       | SMS package quota                   |
| `false`      | Wallet (250 LAK × segments) — default |

---

## Validation

The library rejects invalid input locally **before** calling the API:

- **Phone:** must be exactly `20XXXXXXXX` — 10 digits starting with `20`. No `+856`, no leading `0`.
- **Message:** non-empty (whitespace-only rejected), max **500** characters, and no URLs / domains / short links (the API rejects links).
- **Header:** required for `SendCustom`.

Failures return a sentinel error you can compare with `errors.Is`:

```go
if errors.Is(err, wenovasms.ErrInvalidPhone) {
    // fix the phone number
}
```

---

## Error handling

Errors fall into three categories:

### 1. Local validation (`errors.Is`)

| Sentinel             | Meaning                              |
| -------------------- | ------------------------------------ |
| `ErrInvalidKey`      | `Config.Key` was empty               |
| `ErrInvalidPhone`    | Phone is not `20XXXXXXXX`            |
| `ErrInvalidHeader`   | Custom header was empty              |
| `ErrInvalidMessage`  | Empty, too long, or contains a link  |

### 2. Wenova API errors (`errors.As` → `*wenovasms.Error`)

Carries the business `Code`, `Message`, `Path`, `Method`, and `HTTP` status.

```go
var apiErr *wenovasms.Error
if errors.As(err, &apiErr) {
    switch apiErr.Code {
    case wenovasms.CodePackageInsufficient:
        // buy more SMS package quota
    case wenovasms.CodeWalletInsufficient:
        // top up your wallet
    case wenovasms.CodeRateLimited:
        // slow down and retry
    }
}
```

Common codes: `30101` (auth required), `30102` (links not allowed), `30105` (package quota insufficient), `30106` (wallet insufficient), `30307` (invalid token), `30501` (rate limited), `90901` (internal error).

### 3. Gateway results (`errors.As` → `*wenovasms.GatewayError`)

The HTTP request succeeded but the SMS gateway rejected or has not yet confirmed the message (`resultCode != "20000"`). The `SendResult` is returned alongside the error so you can inspect `ResultCode` / `ResultDesc`.

```go
res, err := wnv.SendOTP(ctx, "2012345678", "Your OTP is: 123456")
if err != nil {
    var gwErr *wenovasms.GatewayError
    if errors.As(err, &gwErr) {
        fmt.Printf("gateway: %s %s\n", gwErr.ResultCode, gwErr.ResultDesc)
    }
}
```

> **Note:** `20000` is success. An empty `resultCode` means the gateway hasn't returned a result yet — for delayed/scheduled messages, wait before treating it as failed.

---

## Security

- **Keep your API token on the server.** Never ship it in frontend JavaScript, mobile apps, or public repos.
- The token is only ever placed in the API request body — it is never included in error messages or logs.
- Each method takes a `context.Context`, so you control timeouts and cancellation.

---

## Roadmap

- [ ] `scriptId` authentication support
- [ ] `POST /sms/check-error` transaction status check
- [ ] Lao / Unicode segment counting helper
