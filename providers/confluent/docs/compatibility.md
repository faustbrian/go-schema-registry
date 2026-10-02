# Confluent compatibility policy

Provider v2 uses the published root v2.0.0 contract. Migrate the provider,
root and format imports together; v1 values and sentinels cannot substitute
for v2 values. The major changes nominal Go identity, not wire identity.

The [specification decision register](specification-decisions.md) defines the
supported 8.3.1 service, compatibility-mode, identity, and wire behavior.
Confluent-compatible products require separate pins and evidence; compatibility
with the public Go API does not imply provider or wire equivalence.
