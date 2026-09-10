# schema-registry

[![CI](https://github.com/faustbrian/go-schema-registry/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-schema-registry/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-schema-registry/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Mutation](https://img.shields.io/badge/mutation-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-checked_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-schema-registry.svg)](https://pkg.go.dev/github.com/faustbrian/go-schema-registry)
[![Release](https://img.shields.io/github/v/release/faustbrian/go-schema-registry?sort=semver)](https://github.com/faustbrian/go-schema-registry/releases)
[![Go](https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`schema-registry` provides provider-neutral contracts for explicit schema
registration, resolution, compatibility, bounded caching, offline bundles, and
wire integration. It preserves provider identity and lifecycle differences:
portable SHA-256 fingerprints never stand in for Confluent IDs or AWS Glue
schema-version UUIDs.

Observable specification and provider choices are maintained in the
[specification decision register](docs/specification-decisions.md).

The core module has no implicit registry client. Provider adapters are separate
modules under `providers/`; format adapters are explicit dependencies under
`formats/`.

The root module is stable at v1, requires Go 1.27.0 or newer, and follows
Semantic Versioning.

## Install

```sh
go get github.com/faustbrian/go-schema-registry@v1
```

Import the canonical provider-neutral package directly:

```go
import schemaregistry "github.com/faustbrian/go-schema-registry"
```

For shared package families, selection guidance, ownership, and lifecycle
vocabulary, see the versioned [v1.4.0 Golib ecosystem
index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Protocols and descriptions family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

## Quick start

```go
adapter, err := registryjsonschema.New(registryjsonschema.Config{
    Dialect:             registryjsonschema.Draft202012,
    MaxSchemaBytes:      64 << 10,
    MaxTotalSchemaBytes: 256 << 10,
    MaxPayloadBytes:     1 << 20,
    MaxResources:        32,
})
if err != nil {
    return err
}

schema, err := schemaregistry.Compile(ctx, schemaregistry.Definition{
    Format:  schemaregistry.FormatJSONSchema,
    Content: rawSchema,
}, adapter)
if err != nil {
    return err
}
fmt.Println(schema.Fingerprint())
```

The complete compiler-checked form is [`ExampleCompile`](example_test.go).

Construct a `Client` with explicit byte, listing, and concurrency limits. Use a
provider adapter only when network operations are intended. Decoding is split
into parse, resolve, and decode phases, so ordinary value access cannot trigger
hidden network I/O.

## Package map

| Package | Use |
| --- | --- |
| `github.com/faustbrian/go-schema-registry` | Define provider-neutral schema identities, registration, resolution, caching, bundles, and wire composition. |
| `github.com/faustbrian/go-schema-registry/formats/avro` | Canonicalize bounded Avro schemas. |
| `github.com/faustbrian/go-schema-registry/formats/jsonschema` | Compile and canonicalize bounded JSON Schema definitions. |
| `github.com/faustbrian/go-schema-registry/formats/protobuf` | Canonicalize bounded Protobuf schemas and imports. |
| `github.com/faustbrian/go-schema-registry/providers/confluent` | Integrate Confluent-compatible REST identity and version-0 wire formats. |
| `github.com/faustbrian/go-schema-registry/providers/glue` | Integrate AWS Glue Schema Registry identity, lifecycle, and uncompressed header-version-3 framing. |

The root compiler, client, cache, bundles, and codecs start no background work
and own no resource that requires shutdown. Applications own injected
providers, transports, credentials, and their lifecycle.

## Contracts

- [API and identity](docs/api.md)
- [Architecture](docs/architecture.md)
- [Provider matrix](docs/providers.md)
- [Evolution and compatibility](docs/evolution.md)
- [Caching and offline bundles](docs/caching.md)
- [Wire formats](docs/wire-formats.md)
- [Authentication and endpoint policy](docs/authentication.md)
- [Outages and incident operation](docs/operations.md)
- [Migration guidance](docs/migrations.md)
- [Kafka, CloudEvents, HTTP, outbox, queue, and workflow examples](docs/examples.md)
- [Security](docs/security.md)
- [Verification provenance](docs/provenance.md)
- [Conformance and hardening matrix](docs/conformance.md)
- [FAQ](docs/faq.md)
- [Documentation index](docs/README.md)
- [Compatibility policy](COMPATIBILITY.md)
- [Support](SUPPORT.md)
- [Private security reporting](SECURITY.md)
- [Changelog](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)

The minimum supported toolchain is Go 1.27.0. The module follows stable v1 compatibility; see
[CHANGELOG.md](CHANGELOG.md) and [RELEASING.md](RELEASING.md).

`golib repository check` validates the repository contract. `golib check --all`
runs the standard gates for the core and both provider modules, including the
provider-specific wire interoperability checks. The release workflow performs
the additional release validation; the optional AWS live integration remains a
caller-selected check and never runs with ordinary CI.

The release gate also runs bounded leak, fault-injection, race-stress, and soak
exercises for the core and both provider modules. Each Go invocation receives
task-owned disposable build and module caches.

## License

MIT. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

## Documentation

Use the [documentation index](docs/README.md) for package-owned guides,
operational contracts, examples, and maintainer references.
