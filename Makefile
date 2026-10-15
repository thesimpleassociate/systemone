BIN := bin/systemone-local
PLUGIN_BIN := $(HOME)/.claude/plugins/cache/systemone/systemone/0.1.0/bin/systemone-local

.PHONY: build test fmt vet install sync clean

build: $(BIN)

$(BIN): $(shell find . -name '*.go')
	go build -o $(BIN) ./cmd/systemone

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

# bin/systemone is a launcher that downloads the release binary. A local build
# lands in bin/systemone-local, which the launcher prefers when present.
# This target is for using systemone as a plain CLI outside Claude Code.
install: build
	install -m 0755 $(BIN) $(HOME)/.local/bin/systemone

# `claude plugin install` copies the binary into its own cache, so a rebuild
# here does not reach the copy the agent runs. This pushes it across without
# a version bump and reinstall.
sync: build
	@test -d "$(dir $(PLUGIN_BIN))" || { echo "plugin not installed at $(PLUGIN_BIN)"; exit 1; }
	install -m 0755 $(BIN) $(PLUGIN_BIN)
	@echo "updated $(PLUGIN_BIN)"

clean:
	rm -f $(BIN)
