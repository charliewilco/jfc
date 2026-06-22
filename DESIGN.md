# Design

`jfc` serves developers who want one dependable formatter for JSON, TOML, YAML, Markdown, JSONL, JSONC, and experimental XML, CSV, TSV, dotenv, and HCL files.

These formats are the product boundary. A file does not need to be a configuration file, manifest, schema, fixture, or generated artifact to belong; it only needs to be one of the supported formats.

The user model is direct:

- Format one file to stdout while inspecting the result.
- Rewrite one file or a directory tree in place.
- Check a repository in CI and fail when supported files are not formatted.
- Pipe stdin through the formatter when another tool owns file discovery.
- Add one small `jfc.toml` when a repository needs local formatting policy or jfc-specific ignores.

## Configuration Model

Configuration is discovered per target. For a file target, `jfc` starts in that file's directory and walks upward until it finds `jfc.toml`. For stdin, `--stdin-filepath` supplies the pretend path used for both format detection and config discovery. Without `--stdin-filepath`, stdin discovery starts from the current working directory and defaults to JSON formatting.

`--config <path>` disables discovery and uses that config for every target in the invocation, including stdin.

Config files do not merge. In a tree with `repo/jfc.toml`, `repo/packages/a/jfc.toml`, and `repo/packages/b/jfc.toml`, files under `packages/a` use the `packages/a` config, files under `packages/b` use the `packages/b` config, and other files use the repo config.

`ignore = [...]` follows the same nearest-config rule. A child config's ignore array replaces the parent config's jfc-specific ignore array; it does not append to it.

There is no `.jfcignore` file. jfc-specific ignores live in `jfc.toml`. Standard ignore sources such as `.ignore`, `.gitignore`, and `.git/info/exclude` are respected as external ignore inputs.

## Formatter Scope

JSON-family files have the strongest structured controls: key sorting, expansion, spacing, and line endings.

YAML is parsed and emitted through `yaml.v3`. The formatter should preserve semantic data, comments, ordering, anchors, aliases, tags, and scalar intent whenever the parser exposes enough information to do so.

TOML is validated and conservatively normalized. It should preserve comments, order, tables, arrays, and string bodies.

Markdown is intentionally a whitespace normalizer. It may normalize line endings, safe blank-line whitespace, and final newlines. It must not reflow prose, reindent code blocks, or treat Markdown as a document rewrite target.

XML is experimental. Element-only XML may be indented, and explicit empty tags must not be rewritten into self-closing tags or vice versa. XML with CDATA, mixed text content, multiline text elements, `xml:space="preserve"`, or multiline/tabbed attributes must be validated without structural pretty-printing so text semantics and SVG authoring details are not changed.

CSV and TSV are experimental validate-only formats. They should catch malformed rows, quoting, and inconsistent field counts without reserializing records or changing embedded field data.

Dotenv is experimental and supports a documented common core. It may normalize assignment spacing around `=`, including optional `export`, but must not interpret interpolation, escape sequences, or dialect-specific quoting without a new policy and tests.

HCL is experimental and delegated to HashiCorp's HCL parser and formatter. jfc may format `.tf` and `.tfvars` files, but it must not become Terraform-aware beyond HCL syntax formatting.

## Traversal Model

Directory traversal is recursive, but only supported file extensions become formatter targets. Traversal prunes ignored directories early, skips `.git`, and does not follow symlinked directories or symlinked files found during a walk.

Explicit symlinked file arguments are accepted when the link path has a supported extension and the target is a regular file. `--write` updates the target file and preserves the symlink.

## Non-Goals

- No `.jfcignore`.
- No implicit config merging.
- No limiting supported formats to configuration files.
- No prose reflow for Markdown.
- No XML rewriting when CDATA, mixed text content, multiline text elements, `xml:space="preserve"`, or multiline/tabbed attributes are present.
- No CSV/TSV canonical serialization without a separate design for quoting, embedded newlines, and consumer compatibility.
- No dotenv dialect expansion without explicit variant rules and fixtures.
- No Terraform project semantics for HCL.
- No broad package-manager distribution before Go install and GitHub release archives are reliable.
- No GitHub Actions polish as a substitute for boring CLI behavior and clear docs.
