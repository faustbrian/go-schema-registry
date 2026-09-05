# Changelog

## Unreleased

### Changed

- Adopt the schema-v2 Golib cohesion catalog contract, including the provider's
  family, selection, ownership, lifecycle, compatibility, and documentation
  metadata and versioned ecosystem navigation.
- Adopt the root repository's checksum-verified `go-library-tools` v1.3.0
  contract and immutable `6c76f5c670d193ce369a7242d4c634f1117286e9`
  workflow while preserving Glue integration and wire interoperability as
  package-owned checks.
- Refresh checksums for the public core and JSON Schema v1.0.0 archives.

### Documentation

- Complete the provider entry point with exact Go, install, import, platform,
  backend, construction, ownership, lifecycle, error, security,
  troubleshooting, support, and ecosystem-navigation guidance.
- Point the provider README and cohesion catalog navigation directly to the
  immutable v1.4.0 ecosystem index and protocols-and-descriptions family
  guidance.
- Record the behavior-neutral AWS SDK Glue release-list review while retaining
  the pinned v1.152.0 API and v1.1.27 Java SerDe contracts.

- Add the [specification decision register](docs/specification-decisions.md),
  exact AWS API/SerDe pins, machine conformance map, and immutable history:
  `GLUE-DEC-001 sha256:1c819f4c7026b332e59bfbc8c5fafb19427231fdc693fdcabeec74f279291acb`,
  `GLUE-DEC-002 sha256:5fbc603e2b39ba635ff321229e0eb433cfc0e4a7478a76abccffac165ee17e9f`,
  and `GLUE-DEC-003 sha256:d90a4c914fb80ecd1c88e5cdbb33bc5e4ea18edbf3e9868a8ce65e2e4ac6d08c`.
- Add a module documentation index for direct navigation.
## 1.0.0 - 2026-08-25

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-schema-registry/providers/glue` identity while preserving its documented API and behavior.

### Documentation

- Link the package README to package-owned documentation.

- Refresh owned-module checksums so clean local and CI dependency resolution
  uses the canonical monorepo versions.
- Exercise provider, scope, registry, and schema-name validation independently
  and simplify exact-length UUID decoding.
- Reject by-ID responses whose returned schema-version UUID differs from the
  requested provider identity.
- Refresh AWS Smithy Go to 1.27.7; the Glue SDK remains current at 1.152.0.
- Reject malformed UUID registration results and missing or mismatched numeric
  version identities in successful AWS responses.
- Reject unsupported configured canonicalizer formats before AWS I/O.
- Added leak, fault-injection, concurrency stress, and bounded soak release
  gates.
- Make the required integration and conformance gates credential-free through
  the real AWS SDK v2 client and a faithful local Smithy JSON service; retain
  caller-selected AWS verification as a separate read-only live target.

### Added

- Bounded AWS SDK v2 registration and resolution with Glue-specific identity,
  lifecycle, error classification, compatibility limitations, and uncompressed
  header-version-3 wire framing.
- Official AWS Glue Java SerDe v1.1.27 wire interoperability, faithful local
  SDK/service integration, and an optional read-only live-service suite for
  existing AVRO schemas.
