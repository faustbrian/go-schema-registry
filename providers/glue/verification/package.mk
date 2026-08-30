SHELL := /usr/bin/env bash

.PHONY: glue-conformance glue-interoperability

glue-conformance:
	./scripts/check-glue.sh

glue-interoperability:
	./scripts/check-wire-reference.sh
