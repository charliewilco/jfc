swear to god, i thought `jq` just did this, but here we are

# jfc

`jfc` is a production-oriented Go CLI for formatting JSON files with Prettier-style ergonomics, including support for real hard tabs and project-local TOML configuration.

## Install

Install the latest version with Go:

```bash
go install github.com/charliewilco/jfc@latest
```

That will place `jfc` in your Go bin directory, typically:

```bash
$(go env GOPATH)/bin
```

If `GOBIN` is set, Go installs there instead.

Build from source in a local checkout:

```bash
go build -o jfc .
```

Install the built binary into a standard system path:

```bash
go build -o jfc .
install -m 0755 ./jfc /usr/local/bin/jfc
```

Install the man page too:

```bash
install -d /usr/local/share/man/man1
install -m 0644 ./man/jfc.1 /usr/local/share/man/man1/jfc.1
```

## Usage

```bash
jfc file.json
cat file.json | jfc
jfc --write file.json
jfc --check .
jfc --list-different .
jfc --config jfc.toml file.json
```

`jfc` reads from stdin when no file paths are provided or when you pass `-`.

```bash
cat package.json | jfc
jfc - < package.json
```

When stdin should inherit config from a specific project path, use `--stdin-filepath`:

```bash
cat payload.json | jfc --stdin-filepath apps/api/payload.json
```

## Examples

Format a single file to stdout:

```bash
jfc package.json
```

Rewrite every JSON file under the current directory:

```bash
jfc --write .
```

Check formatting in CI without mutating files:

```bash
jfc --check .
```

Use hard tabs and key sorting from an explicit config:

```bash
jfc --config ./config/jfc.toml data/payload.json
```

Before:

```json
{"name":"jfc","scripts":{"test":"go test ./...","build":"go build ./..."}}
```

After with default settings:

```json
{
  "name": "jfc",
  "scripts": {
    "test": "go test ./...",
    "build": "go build ./..."
  }
}
```

## CLI Flags

- `--write`: format files in place
- `--check`: print files that are not formatted and exit `1` if any differ
- `--list-different`: print files that differ and exit `1` if any differ
- `--config <path>`: use an explicit `jfc.toml`
- `--stdin-filepath <path>`: resolve config for stdin as if the input came from that file
- `--help`: print CLI usage

## File Matching

`jfc` accepts:

- Individual `.json` files
- Directories, walked recursively
- Shell globs such as `jfc --check 'fixtures/**/*.json'`
- Stdin via no args or `-`

Non-JSON file arguments are rejected instead of being silently skipped. That is stricter than Prettier, but it is the right behavior for a JSON-only formatter.

## Configuration

`jfc` looks for `jfc.toml` in the current project, walking upward from the file being formatted. You can also pass an explicit config path with `--config`.

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

- `use_tabs`: use hard tabs for indentation
- `tab_width`: visual width for one indent level
- `print_width`: target width used for `auto` expansion decisions
- `trailing_newline`: append a final newline when true
- `sort_keys`: sort object keys lexicographically when true
- `array_expand`: one of `"auto"`, `"always"`, or `"never"`
- `object_expand`: one of `"auto"`, `"always"`, or `"never"`
- `space_after_colon`: render `"key": value` vs `"key":value`
- `space_within_braces`: render `{ "x": 1 }` vs `{"x": 1}`
- `space_within_brackets`: render `[ 1, 2 ]` vs `[1, 2]`
- `end_of_line`: one of `"lf"`, `"crlf"`, or `"cr"`

### Formatting Notes

- `sort_keys = false` preserves input object order
- `auto` expansion tries to keep arrays and objects inline when they fit within `print_width`
- `sort_keys = true` only changes object member order; numeric/string values are preserved as parsed JSON

## Cookbook

### Format only changed JSON files in git

```bash
git diff --name-only -- '*.json' | xargs jfc --write
```

### Keep `package.json` stable but sort machine-generated JSON

Use the default config for human-edited files:

```toml
sort_keys = false
object_expand = "auto"
array_expand = "auto"
```

Use a separate config for generated artifacts:

```toml
sort_keys = true
object_expand = "always"
array_expand = "always"
```

Then run:

```bash
jfc --config ./config/generated-json.toml artifacts/*.json
```

### Normalize line endings for cross-platform repos

```toml
end_of_line = "lf"
trailing_newline = true
```

### Pipe from an editor or another tool but still pick up project config

```bash
cat tmp/response.json | jfc --stdin-filepath apps/api/response.json
```

### Make tabs real tabs, not spaces

```toml
use_tabs = true
tab_width = 4
```

## Behavior

- Formats `.json` files from paths, directories, globs, or stdin
- Supports `--write`, `--check`, and `--list-different`
- Returns exit code `1` when `--check` or `--list-different` finds unformatted files
- Returns exit code `2` for parse, config, IO, or usage errors
- Preserves object key order by default and can sort keys with `sort_keys = true`
- Emits clear parse diagnostics with line and column information

## CI

GitHub Actions CI lives at `.github/workflows/ci.yml` and runs:

- `gofmt -l`
- `go vet ./...`
- `go test ./...`
- `go build ./...`

## Man Page

A manual page is included at [man/jfc.1](man/jfc.1).
