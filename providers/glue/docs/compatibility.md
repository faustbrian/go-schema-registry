# AWS Glue compatibility policy

Provider v2 uses the published root v2.0.0 contract. Migrate the provider,
root and format imports together; v1 values and sentinels cannot substitute
for v2 values. The major changes nominal Go identity, not wire identity.

The [specification decision register](specification-decisions.md) defines the
supported SDK API, UUID/lifecycle, capability, and uncompressed wire behavior.
Faithful local Smithy exchanges and Java wire differentials do not prove a live
AWS deployment; credentialed live evidence remains separately reported.
