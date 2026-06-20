# jfc

swear to god, i thought `jq` just did this, but here we are

`jfc` means either "jesus fucking christ" or "just format correctly".

`jfc` is a global formatter CLI for the project files that are always lying around: JSON, JSONC, JSONL, YAML, TOML, and Markdown. Install one binary, point it at a file or repo, and get predictable formatting without remembering a different tool for every format.

```bash
jfc package.json
jfc Cargo.toml
jfc --write .
jfc --check --diff .
```

It is intentionally conservative. JSON-family files get the strongest structured layout controls. TOML is validated and safely normalized without losing comments or order. Markdown is treated as a safe whitespace normalizer, not a prose rewriter.

## Supported Formats

| Format | Extensions | Notes |
| --- | --- | --- |
| JSON | `.json` | Full structured formatting with object/array expansion controls. |
| JSONC | `.jsonc` | Accepts comments and trailing commas while preserving comments; formatting is delegated to `hujson`. |
| JSONL | `.jsonl`, `.ndjson` | Formats each non-empty line as one inline JSON value. |
| YAML | `.yaml`, `.yml` | Preserves mapping order and comments through `yaml.v3` AST formatting. |
| TOML | `.toml` | Validates TOML and normalizes assignment spacing while preserving comments/order. |
| Markdown | `.md`, `.markdown` | Conservative whitespace normalization; no prose reflow. |

## Install

Install the latest version with Go:

```bash
go install github.com/charliewilco/jfc@latest
```

Download checksummed release archives from GitHub Releases when tagged builds are available. Release archives contain the `jfc` binary and `man/jfc.1` for Darwin, Linux, and Windows on `amd64` and `arm64`.

Build from source in a local checkout:

```bash
go build -o jfc .
```

Install the built binary and man page:

```bash
go build -o jfc .
install -m 0755 ./jfc /usr/local/bin/jfc
install -d /usr/local/share/man/man1
install -m 0644 ./man/jfc.1 /usr/local/share/man/man1/jfc.1
```

## Usage

The default mode formats one file or stdin to stdout:

```bash
jfc file.json
jfc config.toml
cat file.json | jfc
cat config.toml | jfc --stdin-filepath config.toml
```

Use `--write` when you want files changed in place:

```bash
jfc --write package.json
jfc --write Cargo.toml
jfc --write .
```

Use `--check` in CI:

```bash
jfc --check .
jfc --check --diff .
```

Initialize the smallest useful project config when you need local ignores:

```bash
jfc init
```

`jfc` reads from stdin when no file paths are provided or when you pass `-`.
Stdin defaults to JSON for backward compatibility. Use `--stdin-filepath` when stdin should use another format or inherit config from a specific project path:

```bash
cat README.md | jfc --stdin-filepath README.md
cat payload.jsonc | jfc --stdin-filepath config/payload.jsonc
```

`--stdin-filepath` matters because stdin has no extension. It tells `jfc` which parser to use and where to start looking for `jfc.toml`.

## CLI Flags

- `--write`: format files in place
- `--check`: print files that are not formatted and exit `1` if any differ
- `--diff`: print formatting changes as a unified diff and exit `1` if any differ; can be combined with `--check`
- `--list-different`: print files that differ and exit `1` if any differ
- `--config <path>`: use an explicit `jfc.toml`
- `--stdin-filepath <path>`: resolve stdin config and format as if input came from that file
- `--help`: print CLI usage

## Commands

### `jfc init`

Create a minimal `jfc.toml` in the current directory:

```toml
ignore = ["dist", "vendor", "node_modules", "*.generated.*"]
```

`jfc init` refuses to overwrite an existing config. jfc-specific ignore patterns live in `jfc.toml`; there is no separate `.jfcignore` file. The CLI also respects standard external ignore sources such as `.ignore`, `.gitignore`, and `.git/info/exclude`.

## File Matching

`jfc` accepts:

- Individual supported files
- Directories, walked recursively
- Shell globs such as `jfc --check 'fixtures/**/*.jsonc'`
- Stdin via no args or `-`

Directory traversal skips unsupported files, `.git`, paths matched by `ignore` in the nearest `jfc.toml`, and paths matched by standard ignore sources such as `.ignore`, `.gitignore`, and `.git/info/exclude`. Explicit unsupported file arguments are rejected instead of silently ignored.

Traversal does not follow symlinked directories or symlinked files discovered while walking a directory. Explicit symlinked file arguments are accepted when the link path has a supported extension and the target is a file; `--write` updates the target file without replacing the symlink. Other hidden, generated, vendor, or build directories are not skipped unless they are unsupported by file extension or named `.git`.

## Configuration

`jfc` looks for `jfc.toml` by walking upward from each file being formatted. `--config` overrides discovery for all targets. Stdin discovery starts from `--stdin-filepath` when provided, otherwise from the current working directory.

Example:

```toml
ignore = ["dist", "vendor", "node_modules", "*.generated.*"]
```

Full example with every supported option:

```toml
use_tabs = true
tab_width = 2
print_width = 80
trailing_newline = true
sort_keys = false
array_expand = "auto"
object_expand = "auto"
space_after_colon = true
space_within_braces = false
space_within_brackets = false
end_of_line = "lf"
ignore = ["dist", "*.generated.json"]
```

### Config Reference

- `use_tabs`: use hard tabs for JSON indentation. YAML always uses spaces because hard-tab indentation is invalid YAML.
- `tab_width`: visual width for tabs and YAML indentation spaces
- `print_width`: target width used for JSON `auto` expansion decisions
- `trailing_newline`: append a final newline when true
- `sort_keys`: sort JSON, JSONC, and JSONL object keys lexicographically when true
- `array_expand`: JSON array layout, one of `"auto"`, `"always"`, or `"never"`
- `object_expand`: JSON object layout, one of `"auto"`, `"always"`, or `"never"`
- `space_after_colon`: render JSON and JSONL object members as `"key": value` vs `"key":value`
- `space_within_braces`: render inline JSON and JSONL objects as `{ "x": 1 }` vs `{"x": 1}`
- `space_within_brackets`: render inline JSON and JSONL arrays as `[ 1, 2 ]` vs `[1, 2]`
- `end_of_line`: one of `"lf"`, `"crlf"`, or `"cr"`
- `ignore`: jfc-specific glob patterns, resolved relative to the config file, for files that `jfc` should skip; patterns without `/` match any path segment

## Format Notes

- JSON uses jfc's own structured renderer, preserves object key order by default, and can sort keys with `sort_keys = true`.
- JSONC preserves comments and accepts trailing commas through `hujson`; jfc can sort object keys, but JSON layout options do not fully control `hujson` spacing.
- JSONL skips blank lines, reports parse errors with the record line number, and renders each record through the JSON formatter in inline mode.
- YAML is parsed and encoded with `yaml.v3`; jfc controls indentation and output conventions but does not expose a full YAML style system.
- TOML formatting is intentionally conservative: invalid TOML is rejected, assignment spacing is normalized, and comments/order are preserved. It does not rewrite tables, arrays, or prose in comments.
- Markdown formatting is intentionally conservative: line endings, safe blank-line whitespace, and safe final-newline conventions are normalized, but prose and code blocks are not rewrapped or reindented.

## Cookbook

Format one JSON file:

```bash
jfc package.json
```

Format one TOML file:

```bash
jfc Cargo.toml
```

Rewrite one file:

```bash
jfc --write package.json
jfc --write pyproject.toml
```

Format a mixed repo in place:

```bash
jfc --write .
```

Check formatting in CI:

```bash
jfc --check .
```

Show CI-friendly formatting failures with exact changes:

```bash
jfc --check --diff .
```

GitHub Actions example:

```yaml
name: Format

on:
  pull_request:

jobs:
  jfc:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go install github.com/charliewilco/jfc@latest
      - run: jfc --check --diff .
```

Preview formatting changes:

```bash
jfc --diff .
```

Format only changed supported files in git:

```bash
git diff --name-only -- '*.json' '*.jsonc' '*.jsonl' '*.ndjson' '*.yaml' '*.yml' '*.toml' '*.md' '*.markdown' | xargs jfc --write
```

Keep generated JSON deterministic:

```toml
sort_keys = true
object_expand = "always"
array_expand = "always"
```

Normalize line endings for cross-platform repos:

```toml
end_of_line = "lf"
trailing_newline = true
```

## Behavior

- Formats supported files from paths, directories, globs, or stdin
- Skips paths ignored by `jfc.toml`, `.ignore`, `.gitignore`, or `.git/info/exclude`
- Supports `--write`, `--check`, `--check --diff`, `--diff`, and `--list-different`
- Returns exit code `1` when `--check`, `--diff`, or `--list-different` finds unformatted files
- Returns exit code `2` for parse, config, IO, or usage errors
- Emits parse diagnostics from the underlying format parser

## CI

The local verification recipe runs:

```bash
just check
```

which currently executes:

- `gofmt -l .`
- `go vet ./...`
- `go tool gotestsum --format testname -- -count=1 <packages with tests>`
- `go build ./...`

Formatter conformance and release packaging smoke tests are available separately:

```bash
just conformance
just release-check
```

`just release-check` cross-compiles Darwin, Linux, and Windows binaries for `amd64` and `arm64`, packages each binary with the man page, writes `checksums.txt`, and verifies archive contents.

## Performance

Formatter benchmark baselines are available through:

```bash
just bench
```

This reports throughput and allocations for representative JSON, TOML, and Markdown inputs.

## Fuzzing

Longer correctness runs are available through:

```bash
just fuzz
FUZZTIME=30s just fuzz-json
FUZZTIME=30s just fuzz-toml
FUZZTIME=30s just fuzz-markdown
```

The default fuzz duration is `10s` per target.

## Man Page

A manual page is included at [man/jfc.1](man/jfc.1).
