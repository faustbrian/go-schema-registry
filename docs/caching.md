# Caching and offline bundles

[Documentation index](README.md)

`ResolveCache` bounds entries, concurrent loads, positive freshness, stale
eligibility, and negative freshness. A selector has one upstream flight, but
each waiter applies its own policy:

- `FailClosed` returns upstream failure.
- `AllowStale` returns an eligible stale value only for `ErrUnavailable` and
  exposes that outage as `StaleCause`. Authoritative absence, authorization,
  cancellation, malformed responses, and identity mismatches remain failures.
- `CacheOnly` performs no provider I/O and returns `ErrOfflineMiss` without a
  usable entry.
- `ReturnUnavailable` performs no lookup and returns `ErrUnavailable`.

`Prime` is the explicit preload path. Results are verified against their lookup
before storage to prevent cache poisoning. `Prime` and `Invalidate` fence any
older in-flight load for that selector, so stale completion cannot overwrite a
preload or repopulate an invalidated entry. This does not cancel old owners or
retract the result received by their existing waiters. Stale eligibility/age uses
each caller's time sampled before upstream work, not a universal expiry recheck
after a long load. Observers receive only state and outcome, never schema/selector
content; they run synchronously after every path, outside locks. Panics are
contained but a blocking observer can delay return.

`Bundle.MarshalBinary` creates a deterministic versioned local artifact with
provenance. `LoadBundle` applies byte and graph bounds, recompiles every schema
with local canonicalizers, verifies every claimed fingerprint, and rejects
missing dependencies in the root closure. Bundle accessors never fetch remote
resources; supplied load canonicalizers must honor the local-only obligation.
Distribute bundles as immutable release artifacts and verify their outer digest
or signature in the deployment system.

Continue with [API and identity](api.md).
