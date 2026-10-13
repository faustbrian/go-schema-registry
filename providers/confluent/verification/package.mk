GOLIB ?= golib

SHELL := /usr/bin/env bash

.PHONY: confluent-interoperability confluent-coverage-guard

confluent-interoperability:
	./scripts/check-confluent.sh
	./scripts/check-wire-reference.sh

# Native coverage remains visible; this owner rejects every other 64-bit gap.
confluent-coverage-guard:
	go test -count=1 -timeout=30s scripts/check-coverage-exception.go scripts/check-coverage-exception_test.go
	GOLIB="$(GOLIB)" bash scripts/check-coverage-exception.sh
