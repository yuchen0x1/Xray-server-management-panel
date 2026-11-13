# Changelog

All notable changes to this fork are recorded here. Versions follow the fork's
own numbering; tags `v3.9.0`–`v3.11.0` belong to upstream 3x-ui and are
deliberately skipped so the two tag namespaces never collide.
Dates use Asia/Shanghai. User-facing Chinese release notes and the release index
are maintained in [docs/releases/release-notes.md](docs/releases/release-notes.md).

## Unreleased

- Stop bundling the fixed third-party OpenSSL installer in future Windows
  archives; document how to obtain OpenSSL separately. Existing release assets
  are unchanged.
- Keep dependency updates on main: disable automatic update PR creation while
  retaining vulnerability alerts and CI security checks.

## v3.12.1 — 2026-11-13

- Raise the minimum Go version and Docker builder to 1.27.4 to fix the reachable
  net/http server crash vulnerability GO-2026-6859.
- Docker releases consume checksum-verified release archives and verify panel,
  Xray and MTProto binaries on all five platforms. Correct 386/ARMv6 selection
  and ARM runtime filenames; keep source builds available by default.
- Linux updates verify release archives against either configured checksum mirror
  before installation and roll back when the updated service cannot start.
- Sidebar project and community links are actionable, and release archives include
  a checksum manifest. Stable releases become public after all assets are uploaded.

- SQLite database downloads now use `VACUUM INTO` to create a consistent snapshot,
  including committed WAL data, instead of copying a live database file.
- SQLite SQL dumps read schema and data in one transaction and preserve BLOB
  storage types, embedded NUL text, and quoted table names. Newly created dump
  files use owner-only permissions on systems that support POSIX file modes.
- Database restore now requires a filename/overwrite confirmation, rejects empty
  uploads, prevents repeated submissions, and releases the busy state on errors.
  All 13 locales include the new prompts.
- Concurrent restore calls are rejected across service instances before they
  touch staging files or restart services. No schema or API route changes.
- Development prerequisites now match the Go version required by `go.mod`.

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

## v3.8.0 — 2026-11-08

- Add TUIC v5 inbounds and align the quic-go dependency with the ported protocol.
- Show associated host remarks in the inbound list and add a Ctrl+K command
  palette for navigation and common operations.
- Use the light theme by default on new installations; retain theme selection.
- Propagate client traffic persistence failures instead of silently ignoring them.
- Accept the legacy string representation of xdns finalmask configuration.
- Fix TUIC test assumptions around stream close timing and repository signatures.
- Upgrade to Go 1.27.3 and address the dependency and CodeQL findings identified
  at the time. This is a historical result, not a current security guarantee.
- Remove release path filtering that could suppress builds for pushed tags.
- Compare: https://github.com/yuchen0x1/Xray-server-management-panel/compare/v3.7.0...v3.8.0

## v3.7.0 — 2026-11-07

- Add subscription import links for Clash, Sing-box, Hiddify, Surge and Loon,
  plus desktop import entry points selected by subscription format.
- Add Webhook notifications for DingTalk, WeCom, Feishu and Discord, configured
  with XUI_WEBHOOK_URL and XUI_WEBHOOK_SECRET.
- Rate-limit subscription ID probing per source with a sliding window and a
  15-minute block; apply trusted proxy configuration when resolving client IPs.
- Switch between Gitee and GitHub immediately on API 403/429 responses during
  install and update instead of repeatedly retrying the limited endpoint.
- First retained stable release in this fork. Existing upstream functionality
  is part of the baseline, not newly introduced by this release.
- Source: https://github.com/yuchen0x1/Xray-server-management-panel/tree/v3.7.0
