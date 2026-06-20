# jfc

swear to god, i thought `jq` just did this, but here we are

`jfc` means either "jesus fucking christ" or "just format correctly".

It is a Go CLI for conservative, predictable formatting of common project metadata and Markdown whitespace, with the strongest layout controls for JSON-family files. One binary, one nearest `jfc.toml`, predictable output.

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

```bash
jfc file.json
jfc README.md
cat file.json | jfc
jfc --write .
jfc --check .
jfc --check --diff .
jfc --diff .
jfc --list-different .
jfc --config jfc.toml config/app.yaml
```

`jfc` reads from stdin when no file paths are provided or when you pass `-`.
Stdin defaults to JSON for backward compatibility. Use `--stdin-filepath` when stdin should use another format or inherit config from a specific project path:

```bash
cat README.md | jfc --stdin-filepath README.md
cat payload.jsonc | jfc --stdin-filepath config/payload.jsonc
```

## CLI Flags

- `--write`: format files in place
- `--check`: print files that are not formatted and exit `1` if any differ
- `--diff`: print formatting changes as a unified diff and exit `1` if any differ; can be combined with `--check`
- `--list-different`: print files that differ and exit `1` if any differ
- `--config <path>`: use an explicit `jfc.toml`
- `--stdin-filepath <path>`: resolve stdin config and format as if input came from that file
- `--help`: print CLI usage

## File Matching

`jfc` accepts:

- Individual supported files
- Directories, walked recursively
- Shell globs such as `jfc --check 'fixtures/**/*.jsonc'`
- Stdin via no args or `-`

Directory traversal skips unsupported files, `.git`, and paths matched by `ignore` in the nearest `jfc.toml`. Explicit unsupported file arguments are rejected instead of silently ignored.

Traversal does not follow symlinked directories or symlinked files discovered while walking a directory. Explicit symlinked file arguments are accepted when the link path has a supported extension and the target is a file; `--write` updates the target file without replacing the symlink. Other hidden, generated, vendor, or build directories are not skipped unless they are unsupported by file extension or named `.git`.

## Configuration

`jfc` looks for `jfc.toml` by walking upward from each file being formatted. `--config` overrides discovery for all targets. Stdin discovery starts from `--stdin-filepath` when provided, otherwise from the current working directory.

Example:

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
- `ignore`: glob patterns, resolved relative to the config file, for files that `jfc` should skip; patterns without `/` match any path segment

## Format Notes

- JSON uses jfc's own structured renderer, preserves object key order by default, and can sort keys with `sort_keys = true`.
- JSONC preserves comments and accepts trailing commas through `hujson`; jfc can sort object keys, but JSON layout options do not fully control `hujson` spacing.
- JSONL skips blank lines, reports parse errors with the record line number, and renders each record through the JSON formatter in inline mode.
- YAML is parsed and encoded with `yaml.v3`; jfc controls indentation and output conventions but does not expose a full YAML style system.
- TOML formatting is intentionally conservative: invalid TOML is rejected, assignment spacing is normalized, and comments/order are preserved. It does not rewrite tables, arrays, or prose in comments.
- Markdown formatting is intentionally conservative: line endings, safe blank-line whitespace, and safe final-newline conventions are normalized, but prose and code blocks are not rewrapped or reindented.

## Cookbook

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

- `go tool gotestsum --format testname -- -count=1 <packages with tests>`
- `go build ./...`

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
