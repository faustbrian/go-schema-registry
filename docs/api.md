# API and identity

[Documentation index](README.md)

`Definition` is mutable caller input. `Compile` copies it and returns an
immutable `Schema`. A portable fingerprint is SHA-256 over a domain separator,
the explicit format, canonical bytes, and reference name/fingerprint pairs in
name order. Metadata and provider coordinates are excluded. Duplicate reference
names are invalid. A hash match with unequal canonical content is a collision,
not equality.

`ProviderID` is opaque and scoped by provider plus provider-defined scope.
Confluent integer IDs and Glue UUIDs are never portable identities. `Lookup`
constructors prevent mixing ID, fingerprint, subject/version, and latest
selectors. Latest is usable only when advertised by `Capabilities`.
Numeric and opaque versions are mutually exclusive. Versioned requests and
nonempty registration-result versions require the selected provider capability.

`RegisterResult.Outcome` distinguishes created, existing, incompatible,
rejected, unauthorized, unavailable, and unknown. Providers that cannot prove
which concurrent request created a version advertise that limitation and return
unknown rather than guessing.
Successful results must include a provider-scoped ID; created outcomes are
accepted only from providers that advertise a reliable creation distinction.
Resolution accepts declared pending/failed/unknown lifecycle as explicit evidence;
nil error does not prove availability. Core selector checks do not universally
validate every response byte, format or version capability.

Deletion confirmation is a provider obligation, not atomic comparison in Client.
Confluent resolves the fingerprint and then sends a separate DELETE; callers own
uncertain mutation reconciliation, cache invalidation and irreversible effects.

`CompatibilityResult.Supported` must be checked before `Compatible`.
Provider-specific semantics use an explicit mode and diagnostic rather than a
portable label. Candidate format and size use the same client bounds as
registration, and provider-specific mode names cannot accompany portable modes.
## Module identity

The root API on main uses `github.com/faustbrian/go-schema-registry/v3`.
Its format packages share that published module. Confluent source targets its
own v3 module and the root v3 types; provider publication remains separate.
Glue retains the published v2 cohort and must not be mixed with v3 clients. See
[migration guidance](migration-v3.md) for coordinated adoption and errors.
