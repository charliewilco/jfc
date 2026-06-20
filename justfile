# Task runner for local development.
# Run `just` to see the available recipes.

test_packages := `go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...`

# Show the recipe list by default.
default:
	@just --list

# Run the CLI without installing it first.
# Everything after `--` is passed to `go run .`.
run *args:
	go run . {{args}}

# Build a local `./jfc` binary in the repo root.
build:
	go build -o jfc .

# Run the Go test suite.
test:
	go tool gotestsum --format testname -- -count=1 {{test_packages}}

# Check that Go sources are gofmt-formatted.
fmt-check:
	@unformatted="$(gofmt -l .)"; \
	if [ -n "$unformatted" ]; then \
		echo "These files are not gofmt-formatted:"; \
		echo "$unformatted"; \
		exit 1; \
	fi

# Run Go's static analyzer.
vet:
	go vet ./...

# Run formatter conformance suites.
conformance:
	go test ./internal/jfc -run 'TestFormat(JSON|TOML).*Conformance|TestFormatMarkdownPreservesRenderedConformanceCases|TestFormatDocumentFixtures|TestFormatJSONCSortCommentFixture' -count=1

# Run formatter performance benchmarks with allocation stats.
bench:
	go test ./internal/jfc -run '^$' -bench 'BenchmarkFormat' -benchmem -count=1

# Run all formatter fuzz suites. Override duration with `FUZZTIME=30s just fuzz`.
fuzz: fuzz-json fuzz-toml fuzz-markdown

# Run JSON formatter fuzz suites.
fuzz-json:
	go test ./internal/jfc -run '^$' -fuzz FuzzFormatJSONMatchesStrictDecoderAcceptance -fuzztime "${FUZZTIME:-10s}" -parallel 1
	go test ./internal/jfc -run '^$' -fuzz FuzzFormatJSONPreservesSemanticsAndIsIdempotent -fuzztime "${FUZZTIME:-10s}" -parallel 1

# Run TOML formatter fuzz suite.
fuzz-toml:
	go test ./internal/jfc -run '^$' -fuzz FuzzFormatTOMLPreservesSemanticsAndIsIdempotent -fuzztime "${FUZZTIME:-10s}" -parallel 1

# Run Markdown formatter fuzz suite.
fuzz-markdown:
	go test ./internal/jfc -run '^$' -fuzz FuzzFormatMarkdownPreservesRenderedHTMLAndIsIdempotent -fuzztime "${FUZZTIME:-10s}" -parallel 1

# Apply standard Go formatting across the module.
fmt:
	go fmt ./...

# Smoke-test release packaging for every supported target.
release-check:
	tmp="$(mktemp -d)"; \
	trap 'rm -rf "$tmp"' EXIT; \
	DIST_DIR="$tmp" bash scripts/release-check.sh

# Basic verification used before handoff: format, vet, tests, and build.
check: fmt-check vet
	go tool gotestsum --format testname -- -count=1 {{test_packages}}
	go build ./...

# Install `jfc` into your Go bin directory.
install:
	go install .

# Copy the locally built binary into `/usr/local/bin`.
bin-install: build
	install -m 0755 ./jfc /usr/local/bin/jfc

# Install the man page into the standard local manpath.
man-install:
	install -d /usr/local/share/man/man1
	install -m 0644 ./man/jfc.1 /usr/local/share/man/man1/jfc.1
