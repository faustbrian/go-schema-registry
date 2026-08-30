# Schema registry decision conformance

The [specification decision register](../docs/specification-decisions.md),
[source pins](../spec/sources.lock.json), and machine-readable
[conformance map](conformance.json) bind each observable choice to its exact
authority and executable evidence.

## Decision matrix

| Decision | Evidence | Differential status |
| --- | --- | --- |
| SCHEMA-REG-DEC-001: Avro parsing canonical form and bounded input | `TestCanonicalizerUsesAvroParsingCanonicalForm` | not assessed; package resource policy |
| SCHEMA-REG-DEC-002: JSON Schema dialect, local resolution, validation, and JCS identity | `TestCanonicalizerUsesOnlyExplicitLocalReferenceResourcesAndDialect` | not assessed; package offline policy |
| SCHEMA-REG-DEC-003: Protobuf linked descriptor canonicalization | `TestCanonicalizerUsesDeterministicLinkedDescriptors` | not assessed; package identity policy |
| SCHEMA-REG-DEC-004: Portable fingerprints remain distinct from provider identities | `TestCompileFingerprintIncludesPortableReferences` | not assessed; provider identities are intentionally different |

Official language sources prove schema semantics. The package-owned canonical
fingerprint corpus proves only portable identity. Provider service and wire
agreement is recorded by the Confluent and Glue modules and does not upgrade
the core into a provider-compatibility claim.

Avro, JSON Schema, and Protobuf publish normative sources rather than one
cross-language official canonical-fingerprint fixture. The package corpus is
therefore classified as owned conformance evidence, not an official fixture.
