# Confluent interoperability boundary

The maintained peer is franz-go `pkg/sr` v1.8.0 at pinned tag archive digest
`c79b6ec629712397b5f2df652eec9de3eebb9b211603a582afa18ebd911f2c49`.
The provider is Confluent Platform 8.3.1; the official wire reference is the
8.3.1 Java `PrefixSchemaIdSerializer`.

`TestProviderAgainstConfluentAndIndependentClient` compares native provider
identity, registration, lookup, references, all compatibility modes, listing,
deletion, JSON value framing, and Avro/JSON Schema/Protobuf corpora. The wire
script compares classic and Protobuf bytes with the official Java serializer.
The pinned comparisons agree. No disagreement is being used as a vote, and no
result certifies other Confluent-compatible products, extensions, quotas, GUID
header version one, or dialect subsets.
