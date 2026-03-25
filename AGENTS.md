# AGENTS

## Repository Notes
- TTL support is implemented via `database/expire.go` and `DB.expires` (unix milliseconds).
- Lazy expiration runs in `DB.GetEntity` and `execKeys`; expired keys are removed on access.
- Any write clears TTL via `DB.Put*`/`Remove` plus explicit clears in hash write commands.
- AOF persistence for TTL always uses absolute `PEXPIREAT`; `SET` with EX/PX and `SETEX` emit `SET` + `PEXPIREAT`.
- Cluster routing adds TTL commands (`expire`, `pexpire`, `expireat`, `pexpireat`, `ttl`, `pttl`, `persist`, `setex`).
