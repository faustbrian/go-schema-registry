# Documentation

The [specification decision register](specification-decisions.md) and
[conformance matrix](../specification/README.md) record the exact AWS service,
live-evidence, and wire boundaries.

## Getting started

- [Package overview](../README.md)
- [Go package documentation](https://pkg.go.dev/github.com/faustbrian/go-schema-registry/providers/glue)
- [Executable construction example](../example_test.go)

## Testing

The module exports no separate testing-helper package. Inject the narrow `API`
interface and canonicalizers as demonstrated by the [provider
tests](../glue_test.go). The faithful local service suite uses `integration`;
the optional read-only AWS suite uses `liveintegration`.

## Security and compatibility

- [Parent security policy](../../../SECURITY.md)
- [MIT license](../LICENSE)
- [Changelog](../CHANGELOG.md)

## Related packages

- [Parent package documentation](../../../docs/README.md)
