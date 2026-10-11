GOLIB ?= golib

.PHONY: check check-full ci cohesion inventory repository-check specification-check specification-check-online

check:
	$(GOLIB) check --local --all

check-full: specification-check-online
	$(GOLIB) check --all

cohesion:
	$(GOLIB) cohesion check

ci: repository-check cohesion specification-check check

inventory repository-check:
	$(GOLIB) repository check

specification-check:
	$(GOLIB) specification check

specification-check-online:
	$(GOLIB) specification check --online
