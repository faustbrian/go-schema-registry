# Architecture

[Documentation index](README.md)

The core owns values, policy validation, bounded concurrency, registration
single-flight, caching, bundles, reference-graph validation, and codec/framer
composition. Format packages own parsing and canonicalization. Provider modules
own remote identity, endpoint, authentication, retry, lifecycle, compatibility,
and deletion semantics. Wire framers are separate from both business codecs and
registry access.

No package uses global registration or service location. Compiled schema/bundle
accessors and built-in wire framers perform no network I/O. Providers and reference
resolvers explicitly accept context. Local-only canonicalizer/value-codec behavior
is an implementation obligation, not an I/O sandbox: arbitrary injected codecs,
canonicalizers, observers, clocks, credentials, SDK clients and transports remain
trusted callbacks and may block, panic or perform external work.

Locks protect only local maps and flight state. They are not held across
provider I/O or observer callbacks. Synchronous flight leaders avoid orphaned
background work; waiters can cancel independently.

Trusted callback panics are not contained. Registration, listing and cache-load
owners nevertheless release their admission during unwinding. Existing flight
waiters receive `ErrUnavailable` without the panic value; registration waiters
receive `RegistrationUnknown` because an external write may already have occurred.
Cache waiters still apply their own availability policy to the raw load failure.

Continue with [API and identity](api.md).
