# Security

## Source applicability and diagnostic boundary

This model covers the maintained root, format packages, and Confluent and Glue
provider modules on main. Published versions retain the behavior of their
immutable tags; these source changes do not establish a new published release.

The root and format source uses the `/v2` identity. The independently versioned
provider directories still use their v1 core dependency until coordinated
provider migration follows public root v2 publication. Their documented private
diagnostic source changes do not imply those changes exist in published v1 tags.
Cross-major types and error sentinels are distinct; see
[the migration guide](migration-v2.md).

Default returned-error formatting uses fixed operation/category messages rather
than schema, subject, reference, payload, credential, or collaborator diagnostic
text. Existing sentinel and retained-cause classification remains available through
`errors.Is`, `errors.As`, and explicit unwrapping. Those inspection mechanisms
are trusted application interfaces, not safe diagnostic rendering interfaces.
Boundaries that deliberately discarded underlying causes retain that policy;
privacy does not introduce new transport or validation-cause disclosure.

## Conditional residual ownership

| Boundary | Owner and rationale | Mitigation | Review trigger |
| --- | --- | --- | --- |
| Explicit cause inspection and structured provider results | Application logging owner; trusted introspection intentionally retains original errors and provider data. | Log fixed categories by default; redact and bound any deliberately inspected cause or provider field. | A new automatic formatter or telemetry integration renders retained data. |
| Synchronous canonicalizers, codecs, credentials and SDK/transport callbacks | Application collaborator owner; Go cannot forcibly preempt arbitrary callback work with context cancellation. | Supply bounded, context-cooperative collaborators, local-only compilers, and transport/SDK deadlines. | New collaborators, retry behavior, or evidence of work continuing beyond the application's budget. |
| Filesystem/network access and credential scope | Application deployment owner; injected transports and SDK clients remain externally managed resources. | Use approved HTTPS endpoints, endpoint-scoped credentials, refused redirects, and explicit client cleanup and quotas. | Endpoint, authentication, transport, or SDK ownership changes. |
| Trusted configuration and aggregate service load | Application resource owner; per-call byte/graph/cache limits do not establish deployment-wide admission. | Set finite supported limits and separately bound ingress, tenants and total concurrent demand. | Raised limits, new tenants, or changed service-level admission. |

`Limits.MaxConcurrent` bounds both executing provider calls and retained
registration leaders. `ResolveCacheConfig.MaxConcurrent` similarly bounds
upstream calls and all retained load leaders, including older generations
detached by invalidation or priming. Existing same-key flights are joined before
admission is checked; new leaders at capacity return fixed `ErrLimitExceeded`
with an empty result before allocating flight state or calling a provider.
Completion releases admission; canceled waiters do not allocate another owner.
Fresh, negative, cache-only and unavailable-policy fast paths retain their
existing behavior. Overload is not provider unavailability and does not permit
stale fallback. These are per-client/cache owner-count bounds, not aggregate
retained-byte, waiter, tenant or deployment-wide admission limits; applications
own those ingress budgets as described above.

Threats include SSRF, redirect credential forwarding, schema bombs, reference
cycles and explosion, oversized responses and payloads, cache poisoning,
compatibility downgrade, concurrent-registration ambiguity, destructive version
deletion, and sensitive schema leakage.

Controls include explicit HTTPS endpoints, injected transports/SDK clients,
endpoint-scoped credentials, total deadlines, bounded retries/concurrency/body
sizes, compile and graph limits, local-only canonicalizers, selector/result
identity validation, negative-cache expiry, explicit stale policy, immutable
bundle verification, and exact fingerprint confirmation before deletion.
Successful remote responses are accepted only when schema content, provider ID,
lifecycle, selector identity, version representation, and advertised provider
capabilities agree; malformed successes are never cached.

Treat subjects, schema names, definitions, diagnostics, and payloads as
potentially sensitive. Observability should use low-cardinality provider,
operation, outcome, lifecycle, and cache-state fields. Do not allow callers to
construct endpoints from message data. Do not follow registry-supplied URLs or
load `$ref`/imports from the network during decode.

The 2026-08-10 dependency and advisory refresh checks Confluent's current
[trust and security resources](https://www.confluent.io/trust-and-security/),
AWS [security bulletins](https://aws.amazon.com/security/security-bulletins/),
and Go dependency vulnerability database through the release `vuln` gates.
Advisory review is time-bound; it is not a claim that future advisories are
absent.
