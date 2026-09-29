# AGENTS.md — plugin-mise

Standalone plugin repo serving the mise image builder and plan-step verb
(`builder:mise` + `verb:mise`). The plugin is a Go module at
`candy/plugin-mise/` (module path
`github.com/opencharly/plugin-mise/candy/plugin-mise`); the root `charly.yml`
declares `discover: candy` (so the repo is a project and its candies are scanned)
plus the per-distro builder boxes and the `check-mise-<distro>-pod` R10 beds.

Canonical files:

- `candy/plugin-mise/charly.yml` — the `plugin-mise:` candy entity (`plugin:`
  block, `plan:` checks).
- `candy/plugin-mise/` — the Go source: `plugin.go`, `provider.go`,
  `mise_builder.go`, `mise_verb.go`, `schema/mise.cue`, `cmd/serve/main.go`.
- `candy/curl-tar/` — the curl+tar toolchain the builder stages need.
- `candy/mise-consumer/` — the shared consumer whose `mise.toml` drives the
  detection path.
- `charly.yml` — the root manifest (`discover: candy`) + the per-distro builder
  boxes + the `check-mise-<distro>-pod` disposable R10 beds.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.
- `README.md` — user overview only; never agent guidance.

## Load these skills first (R0)

- `/charly-internals:plugin` — the plugin authoring reference: the `plugin:`
  block, the `builder` + `verb` provider classes, the per-plugin CUE-schema
  contract, placement.
- `/charly-image:image` — box composition and the builder vocabulary
  (`external_builder:`), the plugin's user-facing surface.
- `/charly-image:layer` — candy authoring and the plan-step verb catalog.
- `/charly-internals:git-workflow` — before any git/PR action.

## Build / validate / test

- `go build ./...` in `candy/plugin-mise/` — compile the plugin module.
- `go test ./...` in `candy/plugin-mise/` — the plugin's Go tests (the builder
  and verb seams).
- `charly box validate` at the repo root — the structural check (the candy +
  `plugin:` block, CUE schema).
- The merge gate is the **org-wide** `charly/pr-validator` (required check
  `validate / validate`, defined in `opencharly/.github`); this repo has **no**
  per-repo candy gate.
- The live R10 witness is `charly check run check-mise-<distro>-pod` (declared in
  the root `charly.yml`, one per distro charly supports).

## Modify this repo

- Edit the `plugin-mise:` candy entity, the Go source, and `schema/mise.cue`
  **together** — the schema is the single source for the verb's `params/` struct.
- The builder stage is distro-agnostic: it never names a distro; the per-distro
  stage bases are the `*-builder` boxes in the root `charly.yml`. Add a distro by
  adding its stage base + a `check-mise-<distro>-pod` bed, not by branching in Go.

## Landing

Load `/charly-internals:git-workflow` before any git/PR action; it owns the
landing mechanics. The authoritative rulebook is the umbrella `AGENTS.md` in
`opencharly/opencharly` and `charly/AGENTS.md` in the charly repo.
