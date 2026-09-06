# Compatibility Policy

Each releasable directory is an independent stable Go module and follows
Semantic Versioning. The root module uses `vX.Y.Z` tags. Provider modules use
their directory-prefixed `providers/confluent/vX.Y.Z` and
`providers/glue/vX.Y.Z` tags.

All currently released modules are at v1. Incompatible exported API or
documented behavior changes require a new major version; minor and patch
releases remain backward compatible within the documented contract.

Compatibility includes exported Go APIs, error classification, serialization,
protocol behavior, persistence schemas, environment variables, command output,
resource ownership, ordering, retry/idempotency semantics, and documented
defaults. A compile-compatible change can still be behaviorally breaking.

Specification-backed modules MUST NOT diverge from their declared standards.
Ambiguities require documented decisions and stable tests. Deprecated APIs
follow [`DEPRECATION.md`](DEPRECATION.md).

The [specification decision register](docs/specification-decisions.md) is the
compatibility authority for Avro, JSON Schema, Protobuf, provider identity, and
wire-policy interpretations.
