# Upstream review history

## 2026-11-26: complete monitoring review

Both monitoring entries were retrieved over HTTPS with the specification
checker's user agent. The Confluent Schema Registry v8.3.1 source archive
remains byte-identical to its monitoring pin. The thirty-day cadence and both
authority URLs remain unchanged.

The tags feed changed and was stable across two consecutive requests:

| Authority | Previous SHA-256 | Current SHA-256 |
| --- | --- | --- |
| `confluent-releases` | `89b2638296f632855f2af8c50c7cb563348c5d262e72617cc3ff429e12ce6c4a` | `99a5e312d34e746fd7f8735755d5d8fb48f902644468046fcb57526bda29ca8a` |

Since the 2026-10-05 review, Confluent published v8.4.0 (2026-10-16), v8.4.1
and v8.3.3 (2026-11-02), and the feed now lists continuous build tags such as
v8.5.0-764, v8.4.2-52 and v8.3.4-51. In v8.3.3 and v8.4.1, the version-zero
and version-one magic bytes and the four-byte schema identifier in `SchemaId`
are unchanged, so CONFLUENT-DEC-002 framing is unaffected. The v8.3.3 REST
entities add an optional `createTs` response field and new association and
provenance resources. These are additive; the adapter tolerates unknown
response fields and does not use the new resources, so CONFLUENT-DEC-001 and
CONFLUENT-DEC-003 are unaffected.

The build tags advance several times a day, so this feed digest is expected
to drift again before the next scheduled review. Retain CONFLUENT-DEC-001
through CONFLUENT-DEC-003, the v8.3.1 source pin and all conformance
bindings. This review refreshes the notification digest only and does not
establish a new compatibility claim.

The `confluent-releases` authority now reads the Maven release metadata
for `kafka-schema-registry-client` instead of the GitHub tags feed. The
tags feed gains `v8.5.0-NNN` build tags several times a day, so its pin
went stale within hours of each review; the Maven metadata lists only
published releases (8.4.1 at this review) and is pinned at
`0ccff277bb846a12474099b5010012ed7470f0de79e3a171387073b73255df0c`.
