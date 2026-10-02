# Root v2 migration

The root source uses `github.com/faustbrian/go-schema-registry/v2` and
Go 1.27.0. Release availability is determined by immutable repository tags and
releases, not this source guide. No version-specific source directory is needed.

Update root imports and the `formats/avro`, `formats/jsonschema`, and
`formats/protobuf` imports together. Reconstruct canonicalizers, schemas,
clients, caches, bundles and codec composition using the same major's types.
The v1 and v2 nominal types are not interchangeable. Sentinel names and
classification semantics remain useful, but their values are distinct across
majors: inspect v2 failures with v2 sentinels, not v1 `errors.Is` targets.

Default diagnostics omit private schema, reference and collaborator text.
Existing retained causes remain available through trusted explicit inspection;
do not automatically log an unwrapped cause. `MaxConcurrent` bounds retained
registration and cache-load leaders as well as active work. Same-key requests
still join admitted work; excess distinct leaders return `ErrLimitExceeded`.
See [the security model](security.md) for caller-owned residual boundaries.

This migration does not change literal portable fingerprints, bundle or wire
representations, provider-scoped identifiers, or supported schema dialects.
Historical specification decisions and the v1 API baseline remain unchanged.

## Provider and consumer sequence

The independently versioned `providers/confluent` and `providers/glue` modules
now target their own `/v2` module identities and the actual published root
v2.0.0 dependency in this source. Their `Config`, provider operations and
framers expose v2 core types. Provider v2 publication and clean consumption
remain separate release steps. Published provider v1 modules still expose v1
types and cannot implement the v2 client contract; do not bridge them by a
cast or alias.

Publish and verify the root v2 release first. Migrate each provider's imports,
module identity and dependency to the actual public root v2 release, then
publish and verify the provider v2 releases. Next, migrate and publish major 2
of the CloudEvents schema-registry adapter in its existing source directory;
its exported nominal types require that owned-adapter major, not an unrelated
CloudEvents root major. Updating only a dependency version cannot bridge majors.
Finally, update and verify Tools' maintained consumer selections against those
actual public releases. Historical v1 consumers and compatibility cohorts retain
their original identities.
