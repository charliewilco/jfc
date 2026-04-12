# swear to god, i thought `jq` just did this, but here we are

# jfc

`jfc` is a production-oriented Go CLI for formatting JSON files with Prettier-style ergonomics, including support for real hard tabs and project-local TOML configuration.

## Install

```bash
go build -o jfc .
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

## Behavior

- Formats `.json` files from paths, directories, globs, or stdin
- Supports `--write`, `--check`, and `--list-different`
- Returns exit code `1` when `--check` or `--list-different` finds unformatted files
- Returns exit code `2` for parse, config, IO, or usage errors
- Preserves object key order by default and can sort keys with `sort_keys = true`
- Emits clear parse diagnostics with line and column information
