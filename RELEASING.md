# Releasing

1. Run `make check-full`, then `golib release check --module .` and
   `golib release dry-run --module .` for the selected module with provider
   integration services explicitly configured where required.
2. Confirm exact statement coverage and viable mutation kills for the core and
   each provider module.
3. Review provider interoperability artifacts, dependency licenses,
   vulnerabilities, SBOM, provenance, benchmarks, API baseline, and the
   Unreleased changelog.
4. After explicit release authorization, create and publish the intended root
   semantic-version tag `vX.Y.Z` from qualified clean main.
5. Release provider modules independently with their own directory-prefixed
   tags. A core release does not imply provider support.

Replace `.` with the provider directory for its metadata check and rehearsal.
Ordinary PR, main and scheduled shared CI are bounded; the explicit full route
and selected release dry-run retain release qualification. These checks do not
publish tags or assets. Publication and actual clean public-consumer verification
remain separate deliberate release operations.
