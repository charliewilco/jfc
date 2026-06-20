# Repository Guidelines

## Project Structure & Module Organization

`jfc` is a Go CLI rooted at `main.go`. Core implementation lives in `internal/jfc`, with format-specific files such as `format_jsonc.go`, `format_yaml.go`, and `format_markdown.go`. Tests sit beside the package code as `*_test.go`. Golden fixtures and sample inputs live under `internal/jfc/testdata/format`. The manual page is maintained at `man/jfc.1`, and common local workflows are defined in `justfile`.

## Build, Test, and Development Commands

- `just` lists available development recipes.
- `just run -- <args>` runs the CLI with `go run .`; for example, `just run -- --check README.md`.
- `just build` builds a local `./jfc` binary.
- `just test` runs `go test ./...`.
- `just fmt` applies `go fmt ./...`.
- `just check` runs the pre-handoff verification path: tests plus `go build ./...`.
- `just install` installs the CLI into your Go bin directory.

## Coding Style & Naming Conventions

Use idiomatic Go and keep code formatted with `gofmt`/`go fmt`; this means tabs for indentation where Go uses indentation. Keep package code in `internal/jfc` unless adding a new public entrypoint. Prefer small, format-focused files and explicit names such as `format_toml.go` or `parser_test.go`. Test functions should follow Go naming conventions, e.g. `TestFormatJSONRejectsInvalidUTF8`.

## Testing Guidelines

This repository uses the standard Go `testing` package. Add tests beside the code they cover and use `t.Parallel()` for independent cases. For formatter behavior, prefer golden fixtures in `internal/jfc/testdata/format` when output readability matters. Run `just test` during development and `just check` before handing off non-documentation changes.

## Commit & Pull Request Guidelines

Recent history uses concise, imperative commit subjects such as `Correct JSONL config documentation` and `Reject table-shaped config values`. Keep commits narrowly scoped. Pull requests should describe the behavior change, mention related issues when applicable, and include the exact validation run, usually `just check`. Include before/after examples for formatter output changes.

## Security & Configuration Tips

Be careful with path traversal, symlink handling, recursive directory walks, and in-place writes. Configuration is discovered through `jfc.toml`; changes to config parsing should cover explicit config files, nearest-file discovery, and stdin behavior through `--stdin-filepath`.
