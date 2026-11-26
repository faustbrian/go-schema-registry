# Upstream review history

## 2026-11-26: complete monitoring review

All eleven monitoring entries were retrieved over HTTPS with the
specification checker's user agent. Apache Avro 1.12.0 and its tags feed,
the JSON Schema Draft 2020-12 Core and Validation source archive, RFC 8785
and its errata, the Protocol Buffers v33.4 source archive, and the
go-schema-registry v1.0.0 source archive remain byte-identical to their
monitoring pins. The thirty-day cadence and all authority URLs remain
unchanged.

Three release feeds changed and were stable across two independent requests:

| Authority | Previous SHA-256 | Current SHA-256 |
| --- | --- | --- |
| `json-schema-releases` | `b56e2044dac3788461a479141eb27334a3bc98354cd8dbfe87998d1ed5e1e9ca` | `0d5c9f06a0d4ec8b04a79fe093ddb895abc08e67481dcb386a6daa29651fc88c` |
| `protobuf-releases` | `894c048c26bf9511be20a7a5e07ef89c277f04d775ccc98f60c2f304e4d226f8` | `90d4e70112cfca01a4c71ada9a444addf1dc886156da179c553684067da7237c` |
| `portable-contract-releases` | `0b5514de392cee3e9969d5249196b45dc16cd22ab5274150eb59b1f1a067b05c` | `d56997c84e6dac9b44d902e2f510635035fd947491314eafcd031aa44494368b` |

Since the 2026-10-05 review, the JSON Schema `main` branch received
editorial and next-draft commits, including regular-expression
clarifications, a schema-evaluation section move, output-schema and format
registry corrections, and the removal of RFC 9557 suffix support from the
`date-time` format (2026-11-19). These commits target the unreleased next
draft. The `draft-bhutton-json-schema-01` archive that defines Draft 2020-12
is unchanged, so SCHEMA-REG-DEC-002, its selected dialect and its JCS
identity rules are unaffected.

Protocol Buffers published v37.0-rc1 (2026-10-15), v37.0-rc2 (2026-11-10)
and v37.0-rc3 (2026-11-12) with matching Objective-C tags. These are release
candidates; the pinned v33.4 language and descriptor sources and the
protocompile v0.14.1 and protobuf-go v1.36.12 dependencies are unchanged, so
SCHEMA-REG-DEC-003 is unaffected.

The go-schema-registry tags feed gained the package's own releases v3.0.0,
v3.0.1 and v3.0.2, Confluent provider v3.0.0 through v3.0.2, and Glue
provider v2.0.1 and v2.1.0. These record this repository's own release
history; the pinned v1.0.0 provider-neutral contract archive is unchanged and
SCHEMA-REG-DEC-004 is unaffected.

Retain SCHEMA-REG-DEC-001 through SCHEMA-REG-DEC-004, all source pins and
conformance bindings. This review refreshes release-feed notification digests
only and does not establish a new compatibility claim.
