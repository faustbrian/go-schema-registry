GOLIB ?= golib

.PHONY: check ci cohesion inventory repository-check

check:
	$(GOLIB) check --all

cohesion:
	$(GOLIB) cohesion check

ci: repository-check cohesion check

inventory repository-check:
	$(GOLIB) repository check
