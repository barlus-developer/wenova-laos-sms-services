# Wenova Link SMS API

Send SMS OTP and notification messages through the Wenova Link REST API.

## Base URL

```text
https://apimicroservices.wenova.fun
```

---

# Send SMS

## `POST /sms/package`

Send an SMS using your SMS package or wallet.

No `Authorization: Bearer` header is required. Authentication is provided using either `token` or `scriptId` in the request body.

### cURL

```bash
curl -X POST "https://apimicroservices.wenova.fun/sms/package" \
  -H "Content-Type: application/json" \
  -d '{
    "header": "WNV-OTP",
    "phoneNumber": "2012345678",
    "message": "Your OTP is: 123456",
    "token": "your-token",
    "usePackage": true
  }'
```

---

## Request Fields

| Field         | Type      |  Required | Description                              |
| ------------- | --------- | --------: | ---------------------------------------- |
| `header`      | `string`  |       Yes | Registered SMS sender ID                 |
| `phoneNumber` | `string`  |       Yes | Lao mobile number in `20XXXXXXXX` format |
| `message`     | `string`  |       Yes | SMS body, maximum 500 characters         |
| `token`       | `string`  | Optional* | API token from Dashboard → Key           |
| `scriptId`    | `number`  | Optional* | Script ID from Dashboard → Key           |
| `usePackage`  | `boolean` |       Yes | `true` = SMS package, `false` = wallet   |

> **Authentication:** At least one of `token` or `scriptId` is required.

---

## `header`

The registered sender ID displayed on the recipient's phone.

Example:

```json
"header": "WNV-OTP"
```

The sender ID must be registered and approved in the Wenova Link dashboard.

---

## `phoneNumber`

The recipient's Lao mobile number.

Required format:

```text
20XXXXXXXX
```

Example:

```json
"phoneNumber": "2012345678"
```

Do **not** include:

```text
+856
```

or a leading `0`.

The API automatically handles the Laos country code when sending to the SMS gateway.

---

## `message`

The SMS message body.

Maximum length:

```text
500 characters
```

Example:

```json
"message": "Your OTP is: 123456"
```

### Message Restrictions

URLs, domains, `www.`, and short links are not allowed.

---

# SMS Segments

SMS messages are billed by segments.

| Message Type         | Characters per Segment |
| -------------------- | ---------------------: |
| English / ASCII      |                    153 |
| Lao / Thai / Unicode |                     67 |

The number of segments is calculated as:

```text
ceil(message length / characters per segment)
```

Minimum: **1 segment**

Maximum message length: **500 characters**

### Example

An English message with 160 characters:

```text
160 / 153 = 2 segments
```

A Lao message with 100 characters:

```text
100 / 67 = 2 segments
```

---

# Authentication

You must provide **at least one** of:

* `token`
* `scriptId`

## Using `token`

Get your API token from:

**Dashboard → Key**

### cURL

```bash
curl -X POST "https://apimicroservices.wenova.fun/sms/package" \
  -H "Content-Type: application/json" \
  -d '{
    "header": "WNV-OTP",
    "phoneNumber": "2012345678",
    "message": "Your OTP is: 123456",
    "token": "your-token",
    "usePackage": true
  }'
```

## Using `scriptId`

Use your Script ID from:

**Dashboard → Key**

### cURL

```bash
curl -X POST "https://apimicroservices.wenova.fun/sms/package" \
  -H "Content-Type: application/json" \
  -d '{
    "header": "WNV-OTP",
    "phoneNumber": "2012345678",
    "message": "Your OTP is: 123456",
    "scriptId": 123,
    "usePackage": true
  }'
```

---

# Payment Source

The `usePackage` field determines how the SMS is charged.

| `usePackage` | Payment Source |
| ------------ | -------------- |
| `true`       | SMS package    |
| `false`      | Wallet         |

## SMS Package

Use:

```json
"usePackage": true
```

SMS is deducted from your active SMS package quota.

The package must:

* Be active
* Be within its start and end dates
* Have enough remaining SMS segments

### cURL

```bash
curl -X POST "https://apimicroservices.wenova.fun/sms/package" \
  -H "Content-Type: application/json" \
  -d '{
    "header": "WNV-OTP",
    "phoneNumber": "2012345678",
    "message": "Your OTP is: 123456",
    "token": "your-token",
    "usePackage": true
  }'
```

## Wallet

Use:

```json
"usePackage": false
```

SMS is charged from your wallet.

Price:

```text
250 LAK × number of segments
```

### cURL

```bash
curl -X POST "https://apimicroservices.wenova.fun/sms/package" \
  -H "Content-Type: application/json" \
  -d '{
    "header": "WNV-OTP",
    "phoneNumber": "2012345678",
    "message": "Your OTP is: 123456",
    "token": "your-token",
    "usePackage": false
  }'
```

> The package or wallet is charged only after the SMS gateway confirms a successful send.

---

# Response

After the HTTP request succeeds, check the SMS gateway response inside `data`.

Important fields:

| Field        | Description                    |
| ------------ | ------------------------------ |
| `resultCode` | SMS gateway result code        |
| `resultDesc` | SMS gateway result description |

A successful SMS normally has:

```text
resultCode: 20000
```

### Example

```json
{
  "data": {
    "resultCode": "20000",
    "resultDesc": "success"
  }
}
```

---

# Check SMS Error

## `POST /sms/check-error`

Check the status of an SMS transaction.

No Bearer token is required.

### cURL

```bash
curl -X POST "https://apimicroservices.wenova.fun/sms/check-error" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "transaction-id",
    "header": "WNV-OTP",
    "phoneNumber": "2012345678",
    "message": "Your OTP is: 123456"
  }'
```

### Request Fields

| Field            | Type     | Required | Description              |
| ---------------- | -------- | -------: | ------------------------ |
| `transaction_id` | `string` |      Yes | SMS transaction ID       |
| `header`         | `string` |      Yes | Registered SMS sender ID |
| `phoneNumber`    | `string` |      Yes | Lao mobile number        |
| `message`        | `string` |      Yes | Original SMS message     |

---

# Error Responses

When a request is rejected, the API returns:

```json
{
  "code": 30101,
  "message": "Either token or scriptId is required",
  "path": "/sms/package",
  "method": "POST"
}
```

## Wenova Business Errors

|    Code | Message                                         | Cause                                         | Solution                               |
| ------: | ----------------------------------------------- | --------------------------------------------- | -------------------------------------- |
| `30101` | Either token or scriptId is required            | Neither authentication field was provided     | Provide `token` or `scriptId`          |
| `30102` | Links are not allowed in SMS messages           | Message contains a URL, domain, or short link | Remove the link                        |
| `30105` | SMS package quota is insufficient               | Package does not have enough segments         | Buy or renew a package                 |
| `30106` | Wallet balance is insufficient for SMS segments | Wallet does not have enough LAK               | Top up your wallet                     |
| `30108` | User account is not active                      | Account has been deactivated                  | Contact support                        |
| `30302` | Wallet not found for user                       | No wallet exists                              | Ensure a wallet exists                 |
| `30303` | SMS package not found                           | No active SMS package                         | Purchase an SMS package                |
| `30307` | SMS API token not found                         | Invalid `token`                               | Get a valid token from Dashboard → Key |
| `30308` | SMS scriptId not found                          | Invalid `scriptId`                            | Get the correct Script ID              |
| `30501` | SMS rate limit exceeded                         | Too many requests                             | Slow down and retry                    |
| `90101` | Request validation failed                       | Invalid request fields                        | Fix the request fields                 |
| `90901` | Internal server error                           | Server-side error                             | Retry later                            |

---

# SMS Gateway Status

The API may return a `resultCode` from the SMS gateway.

| Status         | Meaning                    | Action                                    |
| -------------- | -------------------------- | ----------------------------------------- |
| `20000`        | Success                    | SMS was successfully sent                 |
| Empty / `null` | No result yet              | Check transaction status                  |
| Other          | Gateway rejected or failed | Check `resultDesc` and `developerMessage` |

For delayed or scheduled messages, wait for the gateway result before treating the transaction as failed.

---

# Complete OTP Example

### cURL

```bash
curl -X POST "https://apimicroservices.wenova.fun/sms/package" \
  -H "Content-Type: application/json" \
  -d '{
    "header": "WNV-OTP",
    "phoneNumber": "2012345678",
    "message": "Your OTP is: 123456",
    "token": "your-token",
    "usePackage": true
  }'
```

### Example Variables

```text
Header:       WNV-OTP
Phone:        2012345678
Message:      Your OTP is: 123456
Authentication: token
Payment:      SMS package
```

---

# Security

Keep your API token and Script ID on your **backend/server**.

Do not expose credentials in:

* Frontend JavaScript
* React/Vue/Next.js client-side code
* Mobile application source code
* Public GitHub repositories

Use the SMS API from your backend whenever possible.
