# Changelog

## Unreleased

## 3.0.2 - 2026-10-22

### Changed

- Publish the current source and documentation with a fresh immutable tag.
- Publish independently versioned modules from the same source snapshot:
  - `providers/confluent/v3.0.2`.
  - `providers/glue/v2.1.0`.

## 3.0.1 - 2026-10-13

### Fixed

- Release registration, listing and cache-flight admission when trusted provider,
  resolver or Clock callbacks panic. The original panic still reaches its caller;
  coalesced waiters receive `ErrUnavailable`, and an uncertain registration reports
  `RegistrationUnknown`. No callback preemption or blanket recovery is added.

### Documentation

- Clarify schema/reference identity, copied ownership, actual cache/callback and
  codec admission limits; reconcile SCHEMA-REG-DEC-002/003/004 prose.

### Adoption

- Once both patches are published, select root v3.0.1 and Confluent v3.0.1
  explicitly. The provider retains its root v3.0.0 minimum; selecting only
  the provider patch does not guarantee the root lifecycle repair. Glue v2
  remains a separate root-v2 cohort.

### Source

- Root lifecycle repair and source-contract documentation (other-module
  portions of mixed commits retain their independent release scope):
  [bf563c8fa9](https://github.com/faustbrian/go-schema-registry/commit/bf563c8fa93953c03ff5fa9b25879e35840d88dc),
  [94d4b04fc0](https://github.com/faustbrian/go-schema-registry/commit/94d4b04fc0900cbfa2b17bee4948cb876dad93f2),
  [f61a870d60](https://github.com/faustbrian/go-schema-registry/commit/f61a870d608283b193511cc39c369d7acbce502a),
  [0e75f51f42](https://github.com/faustbrian/go-schema-registry/commit/0e75f51f42d2d5ea8634fff5f2f61dd0c74dc2a1).
- Tooling, documentation enforcement, development policy and bounded
  verification; these are not additional root runtime fixes:
  [8319199b55](https://github.com/faustbrian/go-schema-registry/commit/8319199b5516711d364c324dc0614900bec5a814),
  [0488b1b0dc](https://github.com/faustbrian/go-schema-registry/commit/0488b1b0dc830d1ff91b9f6788bcc2ed37f82522),
  [2aadf27cd0](https://github.com/faustbrian/go-schema-registry/commit/2aadf27cd048c5956c7f38d2cd06462a17f83d73),
  [54cfdd57d5](https://github.com/faustbrian/go-schema-registry/commit/54cfdd57d5f3a8df042feb4c66e5008e348c0eca),
  [84f4d3ed2d](https://github.com/faustbrian/go-schema-registry/commit/84f4d3ed2dcd8513929eef78c21926a310620ea3).
- Already-published Glue v2.0.1 preparation, AWS dependency repair and
  validated-cast explanation; recorded as repository history, not a new Glue
  release or root patch runtime change:
  [516ea3363e](https://github.com/faustbrian/go-schema-registry/commit/516ea3363e8e27c04ff83dae0d7cffc34dbb2268),
  [7d96cac7ec](https://github.com/faustbrian/go-schema-registry/commit/7d96cac7ec2d6fae573b41023f842c6f41736080),
  [14abb8a364](https://github.com/faustbrian/go-schema-registry/commit/14abb8a364e236c2e92c1d9f9ce864d20a1c5a6b).
- Already-published Confluent v3.0.0 adoption and subsequent provider framing
  documentation; these remain provider-scoped, not root runtime fixes:
  [dcc12f1fb9](https://github.com/faustbrian/go-schema-registry/commit/dcc12f1fb90e8352c379d2167d2809b3606b1ca8),
  [b90870929c](https://github.com/faustbrian/go-schema-registry/commit/b90870929ce89397b32815f4944da1b132d1fee9).

## 3.0.0 - 2026-10-08

### Breaking changes

- Move the root and format imports to `github.com/faustbrian/go-schema-registry/v3`
  and adopt JSON Schema v2.0.0. Reconstruct JSON Schema dialect configuration,
  schemas and clients using the same major's types and sentinels. Both existing
  providers remain on their published v2 core cohort; Confluent v3 adoption
  follows root v3 publication.

### Changed

- Accept standard ECMAScript Unicode property patterns. Explicitly asserted
  URI-template formats reject combined prefix/explode modifiers and invalid
  prefix lengths; the default format-annotation policy remains unchanged.

- Update Protocol Buffers to 1.36.12 while preserving deterministic descriptor
  bytes, portable fingerprints, and existing bundle identities.
- Refresh the current Protobuf peer description and retain its prior-runtime
  identity evidence without changing the pinned v33.4 authority.
  SCHEMA-REG-DEC-003 sha256:39202e608558788047e5f9da70faf4cc7d1b53814c5fe844698e5e4245fc1bf0

## 2.0.0 - 2026-10-02

### Security

- Bound retained registration and cache-load leaders by `MaxConcurrent` before
  allocating new flight state. Excess distinct work returns `ErrLimitExceeded`;
  same-key waiters still coalesce, and detached cache loads count until completion.
- Keep collaborator diagnostics and schema/reference identities out of default
  error text while preserving cause inspection, unwrap topology, and bare
  standard cancellation errors at existing passthrough boundaries.

### Changed

- Move the root and its format packages to the `/v2` module identity with
  Go 1.27.0. Applications must migrate the complete root type and sentinel
  graph together; existing provider v1 modules continue using the v1 core.
- Select a separate v2 API projection while preserving the v1 API baseline
  and specification decisions. Literal fingerprint and framing contracts
  are unchanged by the module-path migration.

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

- Complete stable-v1 installation, package selection, lifecycle, executable
  example, performance, compatibility, support, security, and project
  navigation for the root and provider modules.
- Bind root metadata to the compiler-checked example and conformance-based
  performance guide, and clarify root and provider tag formats.

- Point the core README and cohesion catalog navigation directly to the
  immutable v1.4.0 ecosystem index and protocols-and-descriptions family
  guidance.
- Record the behavior-neutral JSON Schema and Protobuf authority reviews,
  including the subsequent JSON Schema example-fence correction, while
  retaining the Draft 2020-12 and Protobuf v33.4 portable contracts and
  refreshing both monitored release-feed digests.
- Record the subsequent behavior-neutral JSON Schema wording and
  validation-output example corrections, Confluent build-feed movement, and
  AWS Glue SDK retry-tracing middleware movement while retaining the immutable
  Draft 2020-12, Platform v8.3.1, and Glue API v1.152.0 contracts and refreshing
  their monitored feed digests.
- Add direct executable provider examples and correct provider documentation,
  test-tag, FAQ, and delivery metadata.

- Add the [specification decision register](docs/specification-decisions.md),
  exact source/update pins, machine conformance map, and immutable decision
  history for the format and portable identity boundaries:
  `SCHEMA-REG-DEC-001 sha256:10a3b137c451307eb5256004208b6693550796fb513305c6cb8022be6b835dce`,
  `SCHEMA-REG-DEC-002 sha256:07d2ac3964542d1a9a91040a7fb53e47255d7ab5c7a054863bacbe9d22c82087`,
  `SCHEMA-REG-DEC-003 sha256:d9353b233c7635b37f288bf42611297f19cfd2603057768bfb65b2a9e2780bfe`,
  and `SCHEMA-REG-DEC-004 sha256:cc435540bc21878a154b9001d5391def7ddebafd09f358a0547f82d2b82fe05b`.
- Replace archived monorepo links and completed execution artifacts with a
  standalone, human-oriented documentation structure.

## 1.0.0 - 2026-08-26

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
