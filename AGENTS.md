# AGENTS.md

## SMS sender headers
- Do not use custom sender headers that start with `WNV-*`.
  Those prefixes are reserved for the built-in defaults
  (`WNV-OTP` for SendOTP, `WNV-info` for SendInfo).
- Custom headers passed to `SendCustom` must not begin with `WNV-`.
