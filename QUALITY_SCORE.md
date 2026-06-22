# Quality Score

Last reviewed: 2026-06-22

Scale: `A` is boringly reliable, `B` is usable with known gaps, `C` needs focused hardening before broad use.

| Area | Grade | Evidence | Next Gap |
| --- | --- | --- | --- |
| User model and docs | A- | README, man page, `DESIGN.md`, and `ARCHITECTURE.md` describe config discovery, stdin, ignores, symlinks, and formatter scope. | Keep examples synchronized with real CLI output as flags change. |
| Config discovery | A | Tests cover nearest root, sibling, child configs, explicit `--config`, stdin discovery, and non-merging ignore replacement. | Add regression cases when config discovery crosses unusual filesystem boundaries. |
| Traversal and writes | B+ | Tests cover ignored-directory pruning, standard ignore pruning, symlink traversal skips, explicit symlink writes, and file mode preservation. | Add more broken-symlink and parent-path edge coverage before changing walk behavior. |
| JSON and JSONL formatting | A- | Structured rendering, key sorting, fixture coverage, fuzz seeds, and semantic assertions cover common project data. | Keep adding real package-lock and generated JSON fixtures when regressions appear. |
| JSONC formatting | B+ | `hujson` preserves comments and trailing commas, with sorting coverage. | Add semantic comparison support for JSONC fixtures if sorting behavior grows. |
| YAML formatting | B | Multi-document streams, comments, anchors, aliases, tags, folded scalars, ordering, and YAML fixtures have coverage. | Continue treating parser presentation loss as high priority; add fixtures from surprising real repos. |
| TOML formatting | A- | TOML syntax validation, comments, ordering, multiline strings, edge fixtures, and fuzzing are covered. | Expand real `Cargo.toml` and `pyproject.toml` fixtures as formatter policy grows. |
| Markdown formatting | A- | Conservative whitespace behavior is covered by fixtures, conformance cases, and fuzzing. | Preserve the no-prose-reflow invariant whenever adding Markdown features. |
| Distribution | B | `release-check.sh` builds archives for supported OS/arch pairs, includes the man page, verifies version output, and writes checksums. | Add release provenance/signing guidance when tagged releases become routine. |
| Agent legibility | A- | `AGENTS.md` is a short map to architecture, design, quality, tests, and release workflow docs. | Promote repeated review comments into tests or docs quickly. |
