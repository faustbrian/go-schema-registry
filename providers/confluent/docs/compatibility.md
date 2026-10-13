# Confluent compatibility policy

[Documentation index](README.md)

Current source targets provider /v3 and published root v3.0.0 types. Existing
provider /v2 releases retain their immutable root /v2 contract; source preparation
does not publish a new provider. Migrate provider, root and format imports together
as described in the [v3 guide](../../../docs/migration-v3.md). Cross-major values
and sentinels are distinct; the nominal module migration retains provider
integer identity and version-0 wire profiles.

The [specification decision register](specification-decisions.md) defines the
supported 8.3.1 service, compatibility-mode, identity, and wire behavior.
Confluent-compatible products require separate pins and evidence; compatibility
with the public Go API does not imply provider or wire equivalence.
