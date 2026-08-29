SHELL := /usr/bin/env bash

.PHONY: confluent-interoperability

confluent-interoperability:
	./scripts/check-confluent.sh
	./scripts/check-wire-reference.sh
