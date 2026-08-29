SHELL := /usr/bin/env bash

.PHONY: glue-interoperability

glue-interoperability:
	./scripts/check-wire-reference.sh
