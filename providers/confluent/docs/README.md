# Documentation

[Package README](../README.md) · [Parent documentation](../../../docs/README.md)

The [specification decision register](specification-decisions.md) and
[conformance matrix](../specification/README.md) record the exact supported
Confluent service and wire boundaries.

## Getting started

- [Package overview](../README.md)
- [Go package documentation](https://pkg.go.dev/github.com/faustbrian/go-schema-registry/providers/confluent/v3)
- [Executable construction example](../example_test.go)

## Testing

The module exports no separate testing-helper package. Inject an
`http.RoundTripper`, `CredentialProvider`, and canonicalizers as demonstrated
by the [provider tests](../confluent_test.go). The optional real-service suite
is isolated behind the separately invoked `confluentintegration` test tag.

## Security and compatibility

- [Parent security policy](../../../SECURITY.md)
- [MIT license](../LICENSE)
- [Changelog](../CHANGELOG.md)

## Related packages

- [Parent package documentation](../../../docs/README.md)

## Ecosystem navigation

See the [versioned Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.8.4/docs/ecosystem/README.md) for package
selection, explicit ownership, and compatibility guidance.

## Additional guides and reference

- [Confluent compatibility policy](compatibility.md)
