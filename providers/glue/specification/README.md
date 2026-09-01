# AWS Glue decision conformance

The [decision register](../docs/specification-decisions.md) and
[source pins](../spec/sources.lock.json) keep API, wire, faithful local, and
credentialed live-service evidence separate.

| Decision | Evidence | Classification |
| --- | --- | --- |
| GLUE-DEC-001: Scoped UUID identity, lifecycle, and unknown registration outcome | `TestProviderPreservesGlueIdentityLifecycleAndUnknownCreationOutcome` | official SDK provider agreement |
| GLUE-DEC-002: Header version three uncompressed wire framing | `TestUncompressedFramerMatchesAWSGlueHeader` | official Java SerDe provider agreement |
| GLUE-DEC-003: Unsupported capabilities and live-service evidence boundary | `TestProviderAgainstFaithfulAWSGlueService` | faithful local; live AWS separately optional |
