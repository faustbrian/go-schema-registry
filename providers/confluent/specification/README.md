# Confluent decision conformance

The [decision register](../docs/specification-decisions.md), [source pins](../spec/sources.lock.json),
and [differential boundary](../spec/interoperability.md) separate service,
maintained-peer, compatibility-corpus, and wire evidence.

| Decision | Evidence | Classification |
| --- | --- | --- |
| CONFLUENT-DEC-001: Scoped identity, registration, references, and lifecycle operations | `TestProviderExposesConfluentSemanticsAndResolvesByGlobalID` | provider and maintained-peer agreement |
| CONFLUENT-DEC-002: Version-zero classic and Protobuf wire framing | `TestClassicFramerMatchesConfluentWireFormatAndBounds` | official provider and maintained-peer agreement |
| CONFLUENT-DEC-003: Compatibility modes and maintained-provider boundary | `TestProviderAgainstConfluentAndIndependentClient` | pinned provider agreement only |
