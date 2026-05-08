.PHONY: workspace ci-local forbidden-symbols build-workspace build-isolated spike clean help

help:
	@echo "Targets:"
	@echo "  workspace          - recreate go.work (gitignored) for local development"
	@echo "  build-workspace    - build all modules in workspace mode"
	@echo "  build-isolated     - build each module with GOWORK=off (matches CI)"
	@echo "  forbidden-symbols  - run the forbidden-symbol grep gate"
	@echo "  ci-local           - run all CI build + grep checks locally"
	@echo "  spike              - run both emission-engine spikes and diff their output (Plan 03)"
	@echo "  clean              - remove go.work, go.work.sum, build artifacts"

workspace:
	go work init
	go work use ./gen ./rt

build-workspace: workspace
	go build ./gen/... ./rt/...

build-isolated:
	cd gen && GOWORK=off go build ./...
	cd rt  && GOWORK=off go build ./...

forbidden-symbols:
	./.github/scripts/forbidden-symbols.sh

ci-local: build-workspace build-isolated forbidden-symbols
	@echo "Local CI checks: OK"

# spike target: Plan 03 owns the spike directory and writes the runner.
# This target shells out to spike/Makefile or runs both engines via shell.
# Behavior: build spike/template and spike/jennifer, run each, diff outputs,
# exit non-zero on any diff. Implementation lives in spike/ to keep this Makefile
# from knowing spike internals.
spike:
	@if [ ! -d spike ]; then echo "spike/ not present (Plan 03 not landed yet)"; exit 0; fi
	@if [ -f spike/run.sh ]; then ./spike/run.sh; \
	 elif [ -f spike/Makefile ]; then $(MAKE) -C spike all; \
	 else echo "spike/ has no runner; see spike/README.md"; exit 1; fi

clean:
	rm -f go.work go.work.sum
	find . -name '*.test' -delete
	find . -name '*.exe' -delete
