# Architecture

`jfc` is a single-binary formatter for JSON, TOML, YAML, Markdown, JSONL, and JSONC. It should stay boring in the best sense: predictable config, conservative formatting, simple traversal, and clear errors when a supported file cannot be handled safely.

This document is the high-level map. Product behavior lives in `DESIGN.md`, agent workflow guidance lives in `AGENTS.md`, and current quality gaps live in `QUALITY_SCORE.md`.

## Codemap

- `main.go` is the thin process entrypoint. It delegates to `internal/jfc.Run`.
- `internal/jfc/app.go` owns CLI parsing, run modes, stdin handling, target collection, directory walking, symlink behavior, and in-place writes.
- `internal/jfc/config.go` owns `jfc.toml` loading, nearest-config discovery, schema validation, and jfc-specific ignore patterns.
- `internal/jfc/ignore_sources.go` owns standard ignore sources: `.ignore`, `.gitignore`, and `.git/info/exclude`.
- `internal/jfc/parser.go`, `document.go`, and `format.go` define supported formats and route inputs to the right formatter.
- `internal/jfc/format_jsonc.go`, `format_jsonl.go`, `format_markdown.go`, `format_toml.go`, and `format_yaml.go` contain format-specific behavior. Keep format policy local to these files when possible.
- `internal/jfc/diff.go` produces unified diffs for check and preview modes.
- `internal/jfc/*_test.go` holds unit, integration, conformance, fuzz seed, fixture, and benchmark coverage beside the package.
- `internal/jfc/testdata/format` contains readable formatter fixtures and golden outputs.
- `scripts/release-check.sh` is the release packaging smoke test for cross-platform archives and checksums.
- `man/jfc.1`, `README.md`, `DESIGN.md`, and `QUALITY_SCORE.md` are the user-facing and maintainer-facing truth surface.

## Execution Flow

`Run` resolves a mode, builds a config loader, expands input paths, collects supported targets, formats each target, and then reports output according to mode.

For stdin, `Run` uses JSON by default. `--stdin-filepath` supplies both format inference and the starting directory for config discovery. `--write` is rejected for stdin.

For files and directories, target collection walks supported files only. The formatter reads the file, formats it in memory, compares bytes, and only writes in `--write` mode when output changed. In-place writes are atomic at the target path and preserve file permissions. Explicit symlinked files write through to the target while preserving the symlink.

## Invariants

- The project config file is `jfc.toml`.
- `--config <path>` is an explicit override for all targets in that invocation, including stdin.
- Without `--config`, each target discovers config by walking upward from that target's directory.
- Nearest config wins. Parent and child configs do not cascade or merge.
- `ignore = [...]` belongs in `jfc.toml`. A nearer config's ignore array replaces any parent config's jfc-specific ignore behavior for targets under the nearer config.
- There is no `.jfcignore` file in this design. Do not add one.
- Standard external ignore sources are supported, but they are not jfc config files.
- Directory traversal skips `.git`, ignored directories, symlinked directories, and symlinked files discovered while walking.
- Explicit unsupported file arguments are errors. Unsupported files found through directory traversal are skipped.
- Format support is format-first, not purpose-first. JSON, TOML, YAML, Markdown, JSONL, and JSONC files are in scope whether or not they are configuration files.
- Markdown formatting must remain conservative and must not reflow prose.
- YAML formatting must treat data loss as the highest-priority bug class.
- Generated release archives must contain the binary and `man/jfc.1`, and `checksums.txt` must cover every archive.

## Cross-Cutting Concerns

Configuration and ignore handling are security-sensitive because they affect which files get read or rewritten. Keep config discovery, ignore bases, and traversal pruning covered by integration tests before changing walk behavior.

Formatter safety is more important than output cleverness. When a format parser preserves comments, anchors, tags, ordering, or scalar style, tests should prove the behavior. When a parser cannot preserve something safely, the formatter should fail loudly or document the boundary.

Distribution should remain simple until the CLI behavior is stable: Go install and GitHub release archives first, package managers later. Signing, provenance, and checksum verification are release-hardening concerns tracked in `QUALITY_SCORE.md`.
