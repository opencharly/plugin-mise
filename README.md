# plugin-mise

The charly plugin serving `builder:mise` + `verb:mise` — full mise (jdx/mise)
support in charly per [PLAN-mise.md](../PLAN-mise.md).

- **builder:mise** — a build-time multi-stage (OpResolve) that provisions the
  mise dev-tool manager into images: distro-aware base, pinned mise release from
  the upstream tarball (Go-style arch mapping), staged MISE_DATA_DIR + shims, and
  `/etc/mise/config.toml` (the system-config path shims resolve against). A candy
  selects it with `external_builder: mise`.
- **verb:mise** — a plan-step verb (`mise: <command>`): OpEmit renders the shell
  the step runs at image build (act-script, wrapped in RUN with the step's
  run_as); OpExecute runs `mise <command>` on the venue at check/deploy time.

Dual-placement by construction: compiled-in (`compiled_plugins:`) or
out-of-process over go-plugin gRPC (`cmd/serve`).

## R10 bed

`charly check run check-mise-pod` (disposable: true) proves the builder + verb
legs end to end: the builder stage provisions mise, the `mise:` verb step
installs node@22, and the checks verify mise + node resolve via the shims.
