# Upstream review history

## 2026-11-26: complete monitoring review

All four monitoring entries were retrieved over HTTPS with the specification
checker's user agent. The AWS SDK Glue v1.152.0 module archive and the AWS
Glue Schema Registry v1.1.27 source archive remain byte-identical to their
monitoring pins. The thirty-day cadence and all authority URLs remain
unchanged.

Two release feeds changed and were stable across two independent requests:

| Authority | Previous SHA-256 | Current SHA-256 |
| --- | --- | --- |
| `aws-glue-api-releases` | `445509ca851d1bc6646a8e3de36b66eeac46af55fe0140ffe8b3bbdbbe31efb9` | `349cfe8ff5c573097d5dbd008301c3a294faf78cb454163956b2349825d91678` |
| `aws-glue-wire-releases` | `afc76371b6e7e5eec462624211572eecb6dbe94bcbff9d3beb686cf7bd879ac5` | `a199a92ee62fba7a92b7bd7c971e1c89a19da0d3951eae55218cc6f885b7df33` |

Since the 2026-10-05 review, the Go module proxy lists
`service/glue` v1.168.0 through v1.177.1. Their changelog covers data
quality, zero-ETL, worker type, job auto-upgrade, Lake Formation function and
business context APIs, smithy-go dependency updates, and two generic
serializer fixes for nil and null list values. No entry changes the schema
registry operations, schema-version identity, lifecycle or modeled errors
that GLUE-DEC-001 and GLUE-DEC-003 rely on.

The AWS Glue Schema Registry SerDe published v2.0.1 (2026-10-07), v2.0.2
(2026-11-09), v3.0.0 and v3.0.1 (2026-11-20). The releases remediate bundled
dependency CVEs and restrict Java class resolution for JSON, Avro and
Protobuf deserialization. `AWSSchemaRegistryConstants` in v3.0.1 still
defines header version byte 3, compression bytes 0 and 5 and a 16-byte schema
version identifier, identical to v1.1.27, so GLUE-DEC-002 framing is
unaffected.

Retain GLUE-DEC-001 through GLUE-DEC-003, all source pins and conformance
bindings. This review refreshes release-feed notification digests only and
does not establish a new compatibility claim.
