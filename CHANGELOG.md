# Changelog

## Unreleased

### Changed

- Adopt the schema-v2 Golib cohesion catalog contract for the core and both
  provider modules, including their family, selection, ownership, lifecycle,
  compatibility, and documentation metadata and versioned ecosystem navigation.
- Adopt the released `go-library-tools` v1.2.0 specification-governance
  contract and retain provider-specific interoperability checks as
  package-owned operations.
- Refresh the `go-json-schema` v1.0.0 checksum after its intentional release
  archive replacement so clean consumers resolve the current archive.
- Adopt the checksum-verified `go-library-tools` v1.3.0 contract and immutable
  `6c76f5c670d193ce369a7242d4c634f1117286e9` workflow while retaining the
  provider-specific interoperability checks.
- Adopt the checksum-verified `go-library-tools` v1.4.0 stable authority-fetch
  contract and immutable `531e4db50fd81a7201257a7b488a0cf22d333aca`
  workflow.

### Documentation

- Point the core README and cohesion catalog navigation directly to the
  immutable v1.4.0 ecosystem index and protocols-and-descriptions family
  guidance.
- Record the behavior-neutral JSON Schema and Protobuf authority reviews,
  including the subsequent JSON Schema example-fence correction, while
  retaining the Draft 2020-12 and Protobuf v33.4 portable contracts and
  refreshing both monitored release-feed digests.

- Add the [specification decision register](docs/specification-decisions.md),
  exact source/update pins, machine conformance map, and immutable decision
  history for the format and portable identity boundaries:
  `SCHEMA-REG-DEC-001 sha256:10a3b137c451307eb5256004208b6693550796fb513305c6cb8022be6b835dce`,
  `SCHEMA-REG-DEC-002 sha256:07d2ac3964542d1a9a91040a7fb53e47255d7ab5c7a054863bacbe9d22c82087`,
  `SCHEMA-REG-DEC-003 sha256:d9353b233c7635b37f288bf42611297f19cfd2603057768bfb65b2a9e2780bfe`,
  and `SCHEMA-REG-DEC-004 sha256:cc435540bc21878a154b9001d5391def7ddebafd09f358a0547f82d2b82fe05b`.
- Replace archived monorepo links and completed execution artifacts with a
  standalone, human-oriented documentation structure.

## 1.0.0 - 2026-08-25

### Changed

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Documentation

- Correct stale package, standalone, and authoritative-source links in public
  documentation.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-schema-registry` identity while preserving its documented API and behavior.

- Make provenance, SBOM, and Confluent wire-reference checks portable to
  minimal CI runners without ripgrep.

### Documentation

- Link the package README to package-owned documentation.

- Replace the obsolete JSON Schema pseudo-version with the canonical monorepo
  dependency version used by clean local and CI resolution.
- Fence cache invalidation and explicit priming against older in-flight loads.
- Reject substituted provider/reference identities, excessive Avro nesting,
  aggregate Protobuf import bytes, and oversized reference or metadata text.
- Refresh the Confluent Platform baseline to 8.3.1, AWS Smithy Go to 1.27.7,
  and the local JSON Schema dependency to its 2026-08-10 revision.
- Added pinned Confluent service, AWS Glue service, independent-client wire,
  provider-module, and clean-consumer release gates.
- Made AWS Glue conformance credential-free with a faithful local Smithy JSON
  service exercised through the real AWS SDK v2 client; live AWS verification
  remains an explicit optional read-only target.
- Added explicit leak, fault-injection, race-stress, soak, provider migration,
  failover, and rollback exercises to the package-local release contract.
- Restricted stale-cache fallback to explicit provider unavailability so
  deletion, authorization, cancellation, and identity failures stay visible.
- Reject ambiguous version identities, incomplete or cross-provider resolution
  results, and provider results that claim unsupported registration semantics.
- Apply format and byte bounds to compatibility candidates, reject contradictory
  compatibility results, and validate administrative lifecycle responses.

### Added

- Provider-neutral immutable schema identities, explicit provider IDs,
  registration outcomes, compatibility results, lifecycle state, references,
  diagnostics, bounded administration, and serializer boundaries.
- Bounded positive and negative resolution caching with per-call outage policy,
  cancellation, single-flight loading, explicit preloading, invalidation, and
  metadata-only observation.
- Verified immutable offline bundles with provenance and portable fingerprints.
- JSON Schema, Avro, and Protobuf canonicalization adapters.
- Separately releasable Confluent-compatible and AWS Glue provider adapters with
  independently versioned wire framers.

No compatibility guarantee applies before the first stable release.
