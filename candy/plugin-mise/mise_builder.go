package mise

// mise_builder.go — the builder:mise legs. Every leg is the ONE shared
// sdk.InvokeBuilder dispatch over sdk/kit (provider.go), exactly like
// candy/plugin-builder-pixi; this file only names the word. The stage is DISTRO-AGNOSTIC: its
// FROM is the host-computed BuilderRef (the image's builder box), the install
// command comes from the embedded builder: mise: vocabulary's install_command,
// and the artifacts copy the Home — no hard-coded distro names, package
// managers, or install paths (the boundary law: distro knowledge is DATA in the
// host's vocabulary, never a builder template).
//
// The builder is selected by DETECTION (a candy shipping mise.toml or
// .tool-versions triggers the embedded builder: mise: vocabulary's detect_file),
// which makes the host compute the FULL BuilderResolveInput (BuilderRef,
// CopySrc, Manifest, UID/GID/Home, InstallCmd, cache mounts).

// builderWord is the reserved builder word this plugin serves.
const builderWord = "mise"
