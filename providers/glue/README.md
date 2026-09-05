# AWS Glue Schema Registry adapter

This independently releasable module preserves AWS Glue registry/schema names,
numeric versions, UUID schema-version IDs, lifecycle state, service errors, and
wire framing. UUIDs remain scoped provider identities, never portable schema
fingerprints.

It owns the bounded adapter from the provider-neutral contract to AWS Glue
registry identity, lifecycle, service, and uncompressed wire semantics. It does
not own AWS region, credentials, endpoint, SDK retry or client shutdown policy,
and it does not advertise references, listing, deletion, candidate
compatibility dry-runs, or ZLIB framing.

The module is stable and active. Its minimum supported and tested toolchain is
Go 1.26.6, matching `go.mod` and the repository manifest. Production source
has no build constraints or operating-system-specific files and supports
portable Go platforms. The faithful local service suite uses the `integration`
test tag; the optional read-only AWS check uses `liveintegration`. The verified
backend and protocol boundary is AWS Glue Schema Registry through AWS SDK for
Go v2 Glue 1.152.0 and uncompressed header-version-3 framing compatible with
AWS Glue Schema Registry Java SerDe 1.1.27.

For shared package families, selection guidance, ownership, and lifecycle
vocabulary, see the versioned [v1.4.0 Golib ecosystem
index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Protocols and descriptions family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

See the [specification decision register](docs/specification-decisions.md) for
the exact API, lifecycle, unsupported-capability, and wire policies.

## Install and import

Install this independently versioned provider module directly:

```sh
go get github.com/faustbrian/go-schema-registry/providers/glue@v1.0.0
```

The canonical import paths and package identifiers are:

```go
import (
	schemaregistry "github.com/faustbrian/go-schema-registry"
	"github.com/faustbrian/go-schema-registry/formats/avro"
	registryglue "github.com/faustbrian/go-schema-registry/providers/glue"
)
```

Use `registryglue` for the provider and framer, `schemaregistry` for the shared
provider-neutral contracts, and an explicit format package such as `avro` for
canonicalization. The alias distinguishes the adapter from the AWS SDK Glue
package. The core contract is the required integration boundary; format
packages are explicit integrations. This module has no nested adapters or
separately owned companion module and does not select a format or codec.

## Quick start

The caller constructs and owns the AWS SDK client:

```go
package main

import (
	"context"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awsglue "github.com/aws/aws-sdk-go-v2/service/glue"
	schemaregistry "github.com/faustbrian/go-schema-registry"
	"github.com/faustbrian/go-schema-registry/formats/avro"
	registryglue "github.com/faustbrian/go-schema-registry/providers/glue"
)

func main() {
	ctx := context.Background()
	aws, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("eu-north-1"))
	if err != nil {
		panic(err)
	}

	provider, err := registryglue.New(registryglue.Config{
		API:            awsglue.NewFromConfig(aws),
		Scope:          "eu-north-1:123456789012:events",
		RequestTimeout: 5 * time.Second,
		MaxConcurrent:  8,
		Canonicalizers: map[schemaregistry.Format]schemaregistry.Canonicalizer{
			schemaregistry.FormatAvro: avro.New(170_000),
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(provider.Capabilities().Provider)
}
```

The provider construction boundary is also compiler-checked and executed as
[`ExampleNew`](example_test.go).

## Construction, ownership, and lifecycle

`Config` receives a narrow AWS SDK v2 Glue client already configured with
region, endpoint, credentials, and SDK retry policy. The adapter adds one total
deadline and a concurrency bound without adding a second retry loop. Use
least-privilege IAM permissions for the exact registry and schema operations.
The caller-provided scope should identify region, account, and registry.

The zero `Config` is invalid; there are no hidden environment, client, retry,
timeout, concurrency, scope, or format defaults. Construction copies the
canonicalizer map and the provider returns copied capability slices. The caller
retains and owns the AWS SDK client and canonicalizers and must keep them valid
for the provider's lifetime.

`Provider` and `UncompressedFramer` start no goroutines and own no AWS client,
connection, or shutdown sequence, so they have no `Close` or `Shutdown`
method. A provider is safe for concurrent use when the supplied AWS SDK client
and every borrowed canonicalizer are safe for concurrent calls;
`MaxConcurrent` bounds admitted service operations. Every potentially
blocking provider operation and each framer method accepts caller context;
service work is additionally bounded by one `RequestTimeout`. The framer checks
cancellation and returns owned frame or payload bytes.

Glue supports Avro, JSON Schema drafts 4/6/7, and proto2/proto3 within service
limits; configure a matching local canonicalizer. Unknown formats are rejected
during construction. The service applies its configured compatibility policy
during registration and does not expose an equivalent candidate dry-run, so
`CheckCompatibility` is explicitly unsupported. The focused adapter does not
advertise schema references, listing, or deletion.

Registration first resolves the exact definition. After registration succeeds,
creation outcome is unknown because duplicate/concurrent calls can return the
same UUID. Resolve exposes pending, available, deleting, failed, or unknown
lifecycle state. The service schema-definition limit is enforced at 170,000
bytes.

## Errors and outcomes

Classify shared failures with `errors.Is` against the provider-neutral
categories, including invalid requests or schemas, unsupported operations,
not found, unavailable, limits, cancellation, deadlines, and unknown outcomes.
AWS modeled and generic service failures are translated at the adapter boundary
while the original SDK operation wrapper and structured API cause remain
available through `errors.Is` and `errors.As`. Non-API SDK failures likewise
retain their original cause chain; public error strings omit cause details.
`ErrInvalidFrame` identifies malformed, cross-scope, or out-of-bounds wire data;
`ErrCompressionUnsupported` identifies recognized ZLIB framing that this
module deliberately does not decode. Registration
returns an unknown outcome when Glue cannot prove which concurrent caller
created a version. Do not retry solely from an error string or add an adapter
retry loop around the SDK retryer; reconcile ambiguous registration first.

## Security and sensitive data

`UncompressedFramer` implements AWS header version 3, compression byte 0, and a
16-byte UUID. ZLIB byte 5 is recognized and explicitly unsupported. See the
official [Glue Schema Registry guide](https://docs.aws.amazon.com/glue/latest/dg/schema-registry.html).

The caller owns region, endpoint, credential rotation, TLS, and SDK retry
policy. Use least-privilege IAM and do not include credentials, schema content,
registry/schema names, payloads, account identifiers, or provider UUIDs in
public diagnostics. The adapter rejects mismatched scopes and identities,
unsupported references and operations, oversized schemas, malformed service
responses, and oversized or compressed frames. See the repository [security
guide](../../docs/security.md) and [private reporting policy](../../SECURITY.md).

## Integration verification

`golib check --module providers/glue` compares framing with the pinned official
AWS Glue Schema Registry Java SerDe v1.1.27 in an isolated Maven container.
Under the `integration` test tag, the same provider contract runs the real AWS
SDK v2 client against a faithful local Smithy JSON service. It requires no AWS
account or credentials and
verifies request serialization, SigV4 wiring, latest/by-ID/version resolution,
registration, duplicates, pending/available lifecycle, SDK throttling retries,
quotas, malformed responses, cancellation, deadlines, unknown outcomes, and
reconciliation.

The `liveintegration` test remains a separate read-only check against a
caller-selected existing AVRO schema using the default AWS credential chain.
It requires `SCHEMA_REGISTRY_GLUE_INTEGRATION_REGION`,
`SCHEMA_REGISTRY_GLUE_INTEGRATION_REGISTRY`, and
`SCHEMA_REGISTRY_GLUE_INTEGRATION_SCHEMA`; it refuses to create a version if
the service cannot find the latest definition. Live access is optional and is
not part of ordinary CI. Shared tooling owns disposable Go caches for
configured checks.

## Documentation and troubleshooting

Use these entry points for the rest of the module contract:

- [Provider compatibility and limitations](docs/compatibility.md)
- [Specification decisions](docs/specification-decisions.md)
- [Provider API](https://pkg.go.dev/github.com/faustbrian/go-schema-registry/providers/glue)
- [Executable construction example](example_test.go)
- [MIT license](LICENSE)
- [Provider comparison and selection](../../docs/providers.md)
- [API and error categories](../../docs/api.md)
- [Operations and troubleshooting](../../docs/operations.md)
- [Migration and provider-switch guidance](../../docs/migrations.md)
- [Composition examples](../../docs/examples.md)
- [FAQ](../../docs/faq.md)
- [Verification and performance evidence](../../docs/conformance.md)
- [Support](../../SUPPORT.md) and [release history](CHANGELOG.md)

The module intentionally exports no separate testing-helper package. Tests can
inject the narrow `API` interface and canonicalizers; the [provider
tests](glue_test.go) demonstrate those deterministic seams.

For construction failures, verify that the SDK client, scope, timeout,
concurrency bound, and supported canonicalizer map are non-zero. For admission
or timeout failures, inspect the caller deadline, SDK retry budget,
`RequestTimeout`, and `MaxConcurrent`. For pending or ambiguous registration,
resolve the returned UUID before treating the schema as available. For frame
failures, confirm header version 3, compression byte 0, UUID syntax and scope,
and the payload bound; compression byte 5 requires a different implementation.

See the [root package documentation](../../README.md) for the provider-neutral
contract and explicit format packages. For ecosystem-wide package selection,
use the immutable [v1.4.0 ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and [Protocols and descriptions guidance](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).
