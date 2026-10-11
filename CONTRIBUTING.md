# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected module's docs.
2. Run focused standalone Go tests for the affected package.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected stable decision entries, and complete the Specification Decisions
section of the pull request template. An unresolved interpretation or stale
source pin is release-blocking; peer behavior cannot silently select policy.

Review the affected decision register before changing observable behavior:
[core and formats](docs/specification-decisions.md),
[Confluent](providers/confluent/docs/specification-decisions.md), or
[AWS Glue](providers/glue/docs/specification-decisions.md).

Explicit full mutation qualification must finish with zero surviving viable
mutants; mutation is not a prerequisite for every ordinary edit.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Use Go 1.27.2 for verification; all public module minimums remain Go 1.27.0.
Standalone untagged tests need neither sibling repositories nor building the
verifier release from source:

```bash
GOWORK=off go test ./... -count=1 -timeout=3m
```

Run the same command from an affected provider module. No goal ledger is a
contributor prerequisite. `make check` and `make ci` use checksum-qualified
public Tools with bounded `--local --all` verification, deterministic Go tests,
saved fuzz seeds and the named race below. Committed-range CI additionally selects
short active fuzz smoke for affected targets; base-free Make commands do not
explore fuzz inputs. No module declares an external
service; Glue retains its `integration` tag, which drives the actual SDK
against a faithful local Smithy service without AWS credentials. This does not
certify a live cloud account or Confluent deployment.

The selected root race runs once within two minutes: a canceled cache waiter
terminates, its leader succeeds, and only one upstream load occurs.

```bash
GOWORK=off go test -race . -run '^TestResolveCacheCoalescesLoadsAndCancelsWaiters$' -count=1 -timeout=2m
```

Normal Go tests execute saved fuzz seeds, not active exploration. Select one
changed target for a two-second active run and optionally one local mutation
package:

```bash
GOWORK=off go test . -run '^$' -fuzz '^FuzzCompileDefinitions$' -fuzztime=2s -timeout=2m
golib mutation --module . --package .
```

Agent-run Go commands must use and remove a task-owned disposable `GOCACHE`.

Shared PR, main and scheduled CI are bounded. Explicit `make check-full` and
the `release_dry_run` input retain full race, coverage, mutation, API/security,
conformance, benchmarks and interoperability. All eight native fuzz targets
retain 100-iteration budgets and two-minute timeouts. Root canonicalization
conformance keeps its five-minute bound; Glue conformance keeps ten minutes;
Confluent conformance and both provider interoperability operations keep
20-minute bounds. All three benchmark operations retain 100ms/20m budgets.
Full Confluent service checks require Docker, curl and its pinned Kafka and
Schema Registry fixtures; both providers' Java wire comparisons require the
pinned Maven container and artifact access. Glue's optional `liveintegration`
check is separate read-only caller-selected AWS access, never ordinary CI.

`make specification-check` is offline; `make specification-check-online` and
full qualification retain online source authority. Use [RELEASING.md](RELEASING.md)
for explicit module-selected rehearsal and deliberate publication. Report every
unavailable or failed selected check without claiming the bounded gate is
release-ready.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
