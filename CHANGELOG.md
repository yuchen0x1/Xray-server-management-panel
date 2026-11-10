# Changelog

All notable changes to this fork are recorded here. Versions follow the fork's
own numbering; tags `v3.9.0`–`v3.11.0` belong to upstream 3x-ui and are
deliberately skipped so the two tag namespaces never collide.

## v3.12.0 — 2026-11-10

Security and reliability release. No breaking configuration changes: every
existing API token keeps working with the privileges it already had.

### API tokens

- Tokens can now carry a **scope**: `admin` (unchanged default), `monitor`
  (read-only status/metrics), `node-sync` (the cluster RPC surface a master
  panel uses), and `node-admin` (node sync plus panel self-update). A token
  presented outside its scope is answered with `403` instead of being treated
  as a full administrator.
- Tokens can carry an **expiry** (`expiresAt`, unix milliseconds). `0` keeps
  the previous never-expires behavior.
- Token lookup is now a single indexed query on the stored SHA-256 hash instead
  of loading every enabled row and comparing in a loop. The comparison still
  uses a constant-time check and no plaintext is ever stored.
- A row whose scope was hand-edited to an unknown value no longer falls back to
  admin privileges; it is rejected.
- Delete and enable/disable accept an optional `expectedScope`, so a stale panel
  cannot revoke a token that was re-minted under the same row id.
- Existing databases are migrated in place: rows created before the scope
  column existed are backfilled to `admin`, so no previously issued token stops
  authenticating.
- The panel UI gained a scope selector, an expiry picker, and a scope badge per
  token.

### API token CLI

- `x-ui setting -getApiToken` now **rotates** a single `cli-fallback` token
  instead of appending a new admin-equivalent token on every invocation, which
  previously left an unbounded pile of credentials nobody could enumerate.

### Two-factor authentication

- Saving settings can no longer silently replace the stored TOTP secret:
  rebinding requires a valid current code (GHSA-xqqw-jqqv-99h6).

### TUIC

- Inbound connections now go through a stateless QUIC Retry and an admission
  check, with pending connections bounded globally and per source /64.
- Pre-authentication stream creation is capped, so an unauthenticated peer can
  no longer open unbounded streams.
- Concurrent sessions and outbound UDP relays are capped per client account,
  and the relay budget is returned when the tunnel closes or the dial fails.
- Fixed a nil-pointer panic in `SocksUDPSession.Close()` when a session was
  closed before both sockets were assigned.

### Subscription server

- The per-source rate limiter no longer drops a new client once its record table
  is full; it evicts expired entries and then the oldest record, so hitting the
  cap can never let an attacker through.
- IPv6 clients are grouped by /64 prefix, so a single host cannot rotate through
  addresses inside its own prefix to bypass the limit.

### Webhooks

- Redirects to a different host are refused, so `X-Webhook-Secret` can no longer
  be replayed to an attacker-controlled redirect target.
- `cpu.high`/`memory.high` are filtered against the webhook's own thresholds
  (`XUI_WEBHOOK_CPU_THRESHOLD` / `XUI_WEBHOOK_MEMORY_THRESHOLD`, default 90)
  instead of firing on every tick.
- In-flight deliveries are bounded; excess events are dropped with a debug log
  instead of spawning an unbounded number of goroutines.
- Alert summaries now report the configured threshold rather than an unset
  value.

### Build

- The bundled Xray binary is pinned to `v26.10.21`, matching the Xray library
  version this build compiles against.
