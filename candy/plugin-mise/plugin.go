// Package mise is the charly plugin serving the `mise` builder + verb: it
// provisions the mise dev-tool manager (jdx/mise) into images via a build-time
// multi-stage (builder:mise, OpResolve) and runs mise commands as plan steps
// (verb:mise — OpEmit at image build, OpExecute at check/deploy). Dual-placement
// by construction: the SAME NewProvider()/NewMeta() compile INTO charly
// in-process when listed in compiled_plugins, or cmd/serve serves them
// OUT-OF-PROCESS over go-plugin gRPC when they are not — placement is invisible
// above the registry.
package mise

import (
	"embed"

	"github.com/opencharly/sdk"
	pb "github.com/opencharly/spec/proto"
)

//go:embed schema/*.cue
var schemaFS embed.FS

// word is the reserved builder + verb word this plugin serves.
const word = "mise"

// NewProvider returns the mise provider (the Invoke dispatch surface).
func NewProvider() pb.ProviderServer { return &provider{} }

// NewMeta advertises builder:mise + verb:mise + the plugin's self-contained CUE
// schema (via sdk.NewMeta → BuildCapabilities). Both capabilities share the ONE
// #MiseInput authoring contract (schema/mise.cue).
func NewMeta() pb.PluginMetaServer {
	return sdk.NewMeta("2026.242.0000",
		[]sdk.ProvidedCapability{
			{Class: "builder", Word: word, InputDef: "#MiseInput"},
			{Class: "verb", Word: word, InputDef: "#MiseInput"},
		},
		schemaFS)
}
