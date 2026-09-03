# Changelog

## Unreleased

### Changed

- Adopt the schema-v2 Golib cohesion catalog contract, including the provider's
  family, selection, ownership, lifecycle, compatibility, and documentation
  metadata and versioned ecosystem navigation.
- Adopt the root repository's checksum-verified `go-library-tools` v1.3.0
  contract and immutable `6c76f5c670d193ce369a7242d4c634f1117286e9`
  workflow while preserving Confluent integration and wire interoperability as
  package-owned checks.
- Refresh checksums for the current core and JSON Schema v1.0.0 archives.

### Documentation

- Add the [specification decision register](docs/specification-decisions.md),
  exact Confluent and peer pins, machine conformance map, and immutable history:
  `CONFLUENT-DEC-001 sha256:67b4c1985e70a7aea45d754e14be846883a37ccd073d76e599d719004af0ea37`, <!-- gitleaks:allow; immutable decision digest -->
  `CONFLUENT-DEC-002 sha256:4c9ab0b72db6bcd6a6f90cd8e638e7f280708c9518128b5933a9a904ad072ff7`, <!-- gitleaks:allow; immutable decision digest -->
  and `CONFLUENT-DEC-003 sha256:c92530ae870914748c82e238b72ff1d3c60c2ec0360150b01863750e6dad0ac1`. <!-- gitleaks:allow; immutable decision digest -->
- Add a module documentation index for direct navigation.
## 1.0.0 - 2026-08-25

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-schema-registry/providers/confluent` identity while preserving its documented API and behavior.

- Make official Java wire-reference verification portable to minimal CI
  runners without ripgrep.

### Documentation

- Link the package README to package-owned documentation.

- Replace the obsolete JSON Schema pseudo-version with the canonical monorepo
  dependency version used by clean local and CI resolution.
- Keep the real Confluent service suite behind its dedicated explicit tag and
  interoperability gate so default tests remain hermetic without weakening
  release proof.
- Validate endpoint policy, deletion coordinates, frame identities, and schema
  references independently while relying on the core schema invariant for
  reference names.
- Reject reference responses whose subject or version does not match the
  requested dependency coordinate.
- Preserve caller cancellation when a retry delay expires concurrently.
- Add byte differentials and equivalent framing benchmarks against Confluent's
  official Java schema-ID serializer 8.3.1.
- Refresh the real-service interoperability baseline to Confluent Platform
  8.3.1 while keeping GUID header wire version 1 explicitly unsupported.
- Check the effective subject or global compatibility policy against complete
  subject history and cover all modes across Avro, JSON Schema, and Protobuf
  service corpora.
- Reject incomplete subject-version results, invalid existing registrations,
  and trailing JSON in otherwise successful registry responses.
- Reject unsupported configured formats and report listed-subject lifecycle as
  unknown because the subject-list response includes no lifecycle evidence.
- Return endpoint-policy errors in conventional lowercase form.
- Added leak, fault-injection, concurrency stress, and bounded soak release
  gates.

### Added

- Bounded Confluent-compatible registration, resolution, references,
  compatibility, listing, deletion, authentication, retries, and version-0
  Avro/JSON/Protobuf framing with explicit provider identity semantics.
- Pinned Confluent Platform 8.3.1 integration coverage with independent
  franz-go identity and wire-format verification.
