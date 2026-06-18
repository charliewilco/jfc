# jfc

swear to god, i thought `jq` just did this, but here we are

`jfc` means either "jesus fucking christ" or "just format correctly".

It is a production-oriented Go CLI for formatting common project metadata and prose files with Prettier-style ergonomics, minus the JavaScript ecosystem. One binary, one nearest `jfc.toml`, predictable output.

## Supported Formats

| Format | Extensions | Notes |
| --- | --- | --- |
| JSON | `.json` | Full structured formatting with object/array expansion controls. |
| JSONC | `.jsonc` | Accepts comments and trailing commas while preserving comments. |
| JSONL | `.jsonl`, `.ndjson` | Formats each non-empty line as one inline JSON value. |
| YAML | `.yaml`, `.yml` | Preserves mapping order and comments through YAML AST formatting. |
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

Directory traversal skips unsupported files and `.git`. Explicit unsupported file arguments are rejected instead of silently ignored.

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
```

### Config Reference

- `use_tabs`: use hard tabs for JSON indentation. YAML always uses spaces because hard-tab indentation is invalid YAML.
- `tab_width`: visual width for tabs and YAML indentation spaces
- `print_width`: target width used for JSON `auto` expansion decisions
- `trailing_newline`: append a final newline when true
- `sort_keys`: sort JSON, JSONC, and JSONL object keys lexicographically when true
- `array_expand`: JSON array layout, one of `"auto"`, `"always"`, or `"never"`
- `object_expand`: JSON object layout, one of `"auto"`, `"always"`, or `"never"`
- `space_after_colon`: render JSON object members as `"key": value` vs `"key":value`
- `space_within_braces`: render inline JSON objects as `{ "x": 1 }` vs `{"x": 1}`
- `space_within_brackets`: render inline JSON arrays as `[ 1, 2 ]` vs `[1, 2]`
- `end_of_line`: one of `"lf"`, `"crlf"`, or `"cr"`

## Format Notes

- JSON preserves object key order by default and can sort keys with `sort_keys = true`.
- JSONC preserves comments and accepts trailing commas.
- JSONL skips blank lines and reports parse errors with the record line number.
- TOML formatting is intentionally conservative: invalid TOML is rejected, assignment spacing is normalized, and comments/order are preserved.
- Markdown formatting is intentionally conservative: line endings, blank-line whitespace, final newline, and fenced code block indentation are normalized, but prose is not rewrapped.

## Cookbook

Format a mixed repo in place:

```bash
jfc --write .
```

Check formatting in CI:

```bash
jfc --check .
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
- Supports `--write`, `--check`, and `--list-different`
- Returns exit code `1` when `--check` or `--list-different` finds unformatted files
- Returns exit code `2` for parse, config, IO, or usage errors
- Emits parse diagnostics from the underlying format parser

## CI

The local verification recipe runs:

```bash
just check
```

which currently executes:

- `go test ./...`
- `go build ./...`

## Man Page

A manual page is included at [man/jfc.1](man/jfc.1).
