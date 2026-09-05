# Confluent-compatible schema registry adapter

This independently releasable module preserves Confluent subject, integer ID,
version, reference, compatibility, and deletion semantics. It does not claim
that every Confluent-compatible service has identical extensions or quotas.

It owns the bounded adapter from the provider-neutral contract to Confluent
REST, subject/version, compatibility, deletion, and version-0 wire semantics.
It does not treat provider IDs as portable fingerprints, discover endpoints,
forward credentials implicitly, own transport shutdown, support GUID header
version 1, or perform unbounded reference traversal.

The module is stable and active. Its minimum supported and tested toolchain is
Go 1.26.6, matching `go.mod` and the repository manifest. It has no build tags
or operating-system-specific source and supports portable Go platforms. The
verified backend and protocol boundary is Confluent Platform 8.3.1, its Schema
Registry REST API, and version-0 classic and Protobuf framing. Other
Confluent-compatible products require their own evidence.

For shared package families, selection guidance, ownership, and lifecycle
vocabulary, see the versioned [v1.4.0 Golib ecosystem
index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Protocols and descriptions family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

See the [specification decision register](docs/specification-decisions.md) for
the exact service, compatibility, and wire policies.

## Install and import

Install this independently versioned provider module directly:

```sh
go get github.com/faustbrian/go-schema-registry/providers/confluent@v1.0.0
```

The canonical import paths and package identifiers are:

```go
import (
	schemaregistry "github.com/faustbrian/go-schema-registry"
	"github.com/faustbrian/go-schema-registry/formats/avro"
	"github.com/faustbrian/go-schema-registry/providers/confluent"
)
```

Use `confluent` for the provider and framers, `schemaregistry` for the shared
provider-neutral contracts, and an explicit format package such as `avro` for
canonicalization. The core contract is the required integration boundary;
format packages are explicit integrations. This module has no nested adapters
or separately owned companion module and does not select a format or codec.

## Quick start

Construction validates policy and performs no network I/O:

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	schemaregistry "github.com/faustbrian/go-schema-registry"
	"github.com/faustbrian/go-schema-registry/formats/avro"
	"github.com/faustbrian/go-schema-registry/providers/confluent"
)

func main() {
	provider, err := confluent.New(confluent.Config{
		Endpoint:         "https://registry.example.com",
		Scope:            "production-cluster",
		Transport:        http.DefaultTransport,
		RequestTimeout:   5 * time.Second,
		MaxResponseBytes: 1 << 20,
		MaxAttempts:      3,
		MaxConcurrent:    8,
		RetryDelay:       100 * time.Millisecond,
		ReferenceLimits: schemaregistry.GraphLimits{
			MaxSchemas: 32, MaxDepth: 8, MaxReferences: 64,
		},
		Canonicalizers: map[schemaregistry.Format]schemaregistry.Canonicalizer{
			schemaregistry.FormatAvro: avro.New(1 << 20),
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(provider.Capabilities().Provider)
}
```

## Construction, ownership, and lifecycle

`Config` requires one HTTPS endpoint, provider scope, injected transport,
request deadline, response/retry/concurrency/reference bounds, and explicit
format canonicalizers. HTTP is available only through the test flag. URL
userinfo, query, fragments, redirects, and implicit credential forwarding are
not accepted. Unknown canonicalizer formats are rejected during construction.

The zero `Config` is invalid; there are no hidden environment, endpoint,
transport, retry, or size defaults. `Credentials` is optional and a zero
`RetryDelay` means no delay. Construction copies the canonicalizer map and the
provider returns copied capability slices. The caller retains and owns the
injected transport, credential provider, and canonicalizers and must keep them
valid for the provider's lifetime.

`Provider`, `ClassicFramer`, and `ProtobufFramer` start no goroutines and own no
connections or shutdown sequence, so they have no `Close` or `Shutdown`
method. A provider is safe for concurrent use when its injected dependencies
are; `MaxConcurrent` bounds admitted HTTP operations. Every potentially
blocking provider operation and each framer method accepts caller context; HTTP
work is additionally bounded by one `RequestTimeout`. Framers check
cancellation and return owned frame or payload bytes.

The adapter retries transport failures, throttling, and server errors within one
total deadline. Registration performs an exact-content lookup first. A
successful create call reports an unknown creation outcome because a concurrent
caller may have created the version. Compatibility is checked only when the
effective subject or global mode matches the requested mode.

`ClassicFramer` implements version-0 Avro/JSON framing. `ProtobufFramer`
implements the version-0 message-index vector. IDs are scoped to the configured
cluster and are not portable fingerprints. Listing returns bounded subject
descriptors with unknown lifecycle because the response does not distinguish
active and soft-deleted state. Soft or permanent version deletion requires an
exact fingerprint confirmation.

## Errors and outcomes

Classify shared failures with `errors.Is` against the provider-neutral
categories, including invalid requests, unsupported operations, not found,
unavailable, confirmation required, reference limits, cancellation, deadlines,
and unknown outcomes. `ErrInvalidResponse` identifies malformed or
identity-inconsistent REST responses; `ErrInvalidFrame` identifies invalid or
out-of-bounds wire data. Registration deliberately returns an unknown outcome
when the service cannot prove which concurrent caller created a version. Do not
retry solely from an error string; preserve the caller's total-work budget and
reconcile ambiguous registration before another write.

## Security and sensitive data

Authentication is supplied by `CredentialProvider` for the configured endpoint.
Use least-privilege credentials, service-specific rate limits, and an endpoint
allowlist. See the upstream [REST API](https://docs.confluent.io/platform/current/schema-registry/develop/api.html)
and [wire format](https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/overview.html).

The caller owns credential rotation and transport TLS policy. Credential
providers must not include authorization values or schema contents in errors;
applications should treat endpoints, subjects, schemas, payloads, and provider
IDs as potentially sensitive. The adapter rejects URL credentials, implicit
redirects, cross-scope IDs, mismatched response identities, excessive response
or reference graphs, and oversized frames. See the repository [security
guide](../../docs/security.md) and [private reporting policy](../../SECURITY.md).

## Integration verification

`golib check --module providers/confluent` starts pinned Confluent Platform 8.3.1 Kafka and Schema
Registry images, runs the adapter against the real REST service, and compares
registration, lookup, listing, references, all compatibility modes across Avro,
JSON Schema, and Protobuf, and classic/Protobuf wire framing with
`franz-go/pkg/sr` v1.8.0 as an independent client. The JSON Schema fixture
also exercises its bounded value codec through a registered schema. Containers,
subjects, and the disposable Go build cache are removed after the run.

The same command compares classic and Protobuf framing byte-for-byte
with Confluent's official Java `PrefixSchemaIdSerializer` from
`kafka-schema-serializer` 8.3.1. It also publishes equivalent 1,024-byte
framing benchmarks for the official serializer and the Go framers. The Maven
runtime, primary Confluent artifact checksum, Maven cache, and Go build cache
are isolated and verified by the gate.

## Documentation and troubleshooting

Use these entry points for the rest of the module contract:

- [Provider compatibility and limitations](docs/compatibility.md)
- [Specification decisions](docs/specification-decisions.md)
- [Provider comparison and selection](../../docs/providers.md)
- [API and error categories](../../docs/api.md)
- [Operations and troubleshooting](../../docs/operations.md)
- [Migration and provider-switch guidance](../../docs/migrations.md)
- [Composition examples](../../docs/examples.md)
- [FAQ](../../docs/faq.md)
- [Verification and performance evidence](../../docs/conformance.md)
- [Support](../../SUPPORT.md) and [release history](CHANGELOG.md)

For HTTP policy failures, first verify that the endpoint is HTTPS and contains
no userinfo, query, or fragment. For admission or timeout failures, verify the
caller deadline, `RequestTimeout`, and `MaxConcurrent`. For compatibility
results, check `Supported` before `Compatible` and confirm the requested mode
matches the effective subject/global mode. For frame failures, confirm the
provider scope, positive integer schema ID, version-0 format, payload bound,
and—when using Protobuf—the message-index bound.

See the [root package documentation](../../README.md) for the provider-neutral
contract and explicit format packages. For ecosystem-wide package selection,
use the immutable [v1.4.0 ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and [Protocols and descriptions guidance](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).
