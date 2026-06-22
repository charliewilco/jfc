# Agent Guide

Start here, then open the smallest deeper document needed for the task.

## Repository Map

- `ARCHITECTURE.md` maps the codebase, execution flow, and invariants.
- `DESIGN.md` explains the user model, config behavior, traversal model, formatter scope, and non-goals.
- `QUALITY_SCORE.md` tracks quality by format family and CLI subsystem over time.
- `README.md` and `man/jfc.1` are the user-facing truth surface.
- `justfile` is the local workflow entrypoint.

## Project Structure

`jfc` is a Go CLI rooted at `main.go`. Core implementation lives in `internal/jfc`, with format-specific files such as `format_jsonc.go`, `format_yaml.go`, `format_markdown.go`, `format_xml.go`, `format_csv.go`, `format_env.go`, and `format_hcl.go`. Tests sit beside package code as `*_test.go`. Golden fixtures and sample inputs live under `internal/jfc/testdata/format`. Release packaging checks live in `scripts/release-check.sh`.

## Commands

- `just` lists available recipes.
- `just run -- <args>` runs the CLI with `go run .`.
- `just build` builds a local `./jfc` binary.
- `just test` runs the Go test suite.
- `just fmt` applies `go fmt ./...`.
- `just check` runs the pre-handoff verification path.
- `just conformance`, `just bench`, `just fuzz`, and `just release-check` cover focused confidence passes.

## Working Rules

- Follow existing project conventions.
- Use idiomatic Go and keep files gofmt-formatted; Go indentation is tabs where gofmt uses indentation.
- Keep package code in `internal/jfc` unless adding a new public entrypoint.
- Prefer focused tests beside the behavior they cover.
- Prefer golden fixtures when formatter output readability matters.
- Documentation-only changes do not require checks unless requested; code changes require relevant checks before handoff.
- Never claim checks passed unless they were actually run.

## Invariants To Protect

- The project config file is `jfc.toml`.
- `--config <path>` overrides discovery for every target in the invocation, including stdin.
- Without `--config`, each target uses nearest-config-wins discovery from that target path.
- Config files do not merge.
- `ignore = [...]` in a nearer config replaces the parent config's jfc-specific ignore list.
- There is no `.jfcignore`; do not add one.
- Standard ignore files remain external ignore sources, not jfc config files.
- Format support is format-first, not purpose-first; supported files are not limited to configuration files.
- XML support is experimental; preserve CDATA, mixed text content, multiline text elements, `xml:space="preserve"`, and multiline/tabbed attributes by falling back to validation-only behavior.
- CSV/TSV support is experimental and validate-only; do not serialize records into a new canonical form without a separate safety design.
- Dotenv support is experimental; keep it to the documented common assignment core unless variant rules and fixtures are added first.
- HCL support is experimental and delegated to HashiCorp tooling; do not add Terraform project semantics.
- Markdown formatting must stay conservative and must not reflow prose.
- YAML data loss is a highest-priority formatter bug class.
- Be careful with path traversal, symlink handling, recursive directory walks, and in-place writes.

## Commit And PR Notes

Recent history uses concise imperative commit subjects such as `Fix install version reporting` and `Prepare first jfc release`. Keep commits narrowly scoped. PRs should describe the behavior change, mention related issues when applicable, and include exact validation, usually `just check`.
