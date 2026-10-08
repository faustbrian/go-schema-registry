# Root v3 migration

The maintained root and format source targets
`github.com/faustbrian/go-schema-registry/v3`, with Go 1.27.0 and JSON Schema
`github.com/faustbrian/go-json-schema/v2` v2.0.0. Publication is a separate step:
verify immutable root and provider tags before selecting them in an application.

Migrate root and format imports together. Reconstruct schemas,
canonicalizers, clients, caches and codecs with the same major's types and
error sentinels. The JSON Schema adapter's `Dialect` alias and `Config.Dialect`
now have JSON Schema v2 type identity; v1 dialect values are not interchangeable.
Both current providers retain their published `/v2` identities and root v2.0.0.
They cannot be supplied to a v3 client. A future Confluent v3 adoption follows
root v3 publication and must migrate its composition using the same major's types.

JSON Schema compilation accepts standard ECMAScript Unicode properties such as
`Script=Greek`. Format remains annotation-only by default. A caller-supplied
meta-schema declaring the format-assertion vocabulary opts into validation;
asserted URI templates reject combined prefix/explode modifiers and invalid
prefix lengths. Resources remain explicit, bounded and caller configured;
compilation does not introduce implicit network retrieval.

Portable fingerprints, provider identities, wire framing, lifecycle ownership,
bounded admission and default diagnostic privacy retain their existing contracts.
Publish and verify root v3 first, then the Confluent v3 provider and consumers
that expose those nominal types. No version-specific source directories are used.

The current Confluent and Glue providers deliberately retain their published
`/v2` modules and root v2.0.0 dependencies. Existing v2 tags, API
projections, specification decisions and compatibility reports remain historical
evidence, not proof of this new cohort. The [v2 guide](migration-v2.md) describes
that earlier cohort.
