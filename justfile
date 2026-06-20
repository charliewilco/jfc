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

# Run formatter performance benchmarks with allocation stats.
bench:
	go test ./internal/jfc -run '^$' -bench 'BenchmarkFormat' -benchmem -count=1

# Apply standard Go formatting across the module.
fmt:
	go fmt ./...

# Basic verification used before handoff: tests plus a build.
check:
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
