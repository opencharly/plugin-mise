# plugin-mise

Full [mise](https://mise.jdx.dev) (jdx/mise) support in OpenCharly — the
`builder:mise` image builder and the `verb:mise` plan-step verb.

The plugin is an out-of-tree Go module served over go-plugin gRPC; the charly
loader host-builds it and serves it out-of-process, and the same provider can be
compiled in.

## What it provides

| Capability | Surface |
|---|---|
| `builder:mise` | a build-time multi-stage (`OpResolve`) that provisions mise into images |
| `verb:mise` | a plan-step verb — `mise: <command>` — `OpEmit` at image build, `OpExecute` at check/deploy |

The builder is distro-aware: it picks the base, installs a pinned mise release
from the upstream tarball (Go-style arch mapping), stages `MISE_DATA_DIR` +
shims, and writes `/etc/mise/config.toml` (the system-config path shims resolve
against). A candy selects it with `external_builder: mise`.

The `mise:` verb's `OpEmit` renders the shell the step runs at image build
(act-script, wrapped in `RUN` with the step's `run_as`); `OpExecute` runs
`mise <command>` on the venue at check/deploy time.

## How to use it

Compose the plugin candy in a box's `candy:` list, select the builder, and author
a step:

```yaml
- check: install go via mise
  mise: install go@1.25
```

Input fields: `command` (also the scalar-sugar primary — `mise: install`),
`args`, `tool`, `task`, `config`, `tools`, `env`, `run_as`.

## R10 beds

`charly check run check-mise-<distro>-pod` (disposable, one per distro charly
supports) proves the builder + verb legs end to end: the builder stage provisions
mise, the detection path installs `zig` from the shared consumer's `mise.toml`,
the `mise:` verb step installs `go@1.25`, and the checks verify mise + the tools
resolve via the shims. The fixture tools ship official libc-agnostic prebuilds
(zig and go are fully static), so the same bed passes on glibc AND musl (Alpine)
bases.

## Layout

- `candy/plugin-mise/` — the plugin module: `plugin.go` (provider + meta),
  `provider.go`, `mise_builder.go`, `mise_verb.go`, `schema/mise.cue` (the
  self-contained `#MiseInput`), and `cmd/serve/main.go`.
- `candy/curl-tar/` — the curl+tar toolchain the builder stages need.
- `candy/mise-consumer/` — the shared consumer whose `mise.toml` drives the
  detection path.
- `charly.yml` — the root project manifest (`discover: candy`) plus the per-distro
  builder boxes and check beds.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.

## Related

- Owning skill: `/charly-image:image` — box composition and the builder
  vocabulary (the plugin candy carries no `skill:` entity of its own; the gap is
  tracked in
  [opencharly/opencharly#291](https://github.com/opencharly/opencharly/issues/291)).
- `/charly-image:layer` — the candy/plan-step authoring surface.
- `/charly-internals:plugin` — the plugin/provider model.
- [`opencharly/charly`](https://github.com/opencharly/charly) — the charly CLI.
