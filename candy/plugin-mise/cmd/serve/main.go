// Command serve is the OUT-OF-PROCESS entrypoint for the mise plugin: a thin
// shim serving the importable provider over go-plugin gRPC via sdk.Serve. The
// SAME NewProvider()/NewMeta() compile INTO charly in-process when listed in
// compiled_plugins; this binary is host-built + connected only when they are
// NOT — placement is invisible above the registry.
package main

import (
	mise "github.com/opencharly/plugin-mise/candy/plugin-mise"
	"github.com/opencharly/sdk"
)

func main() { sdk.Serve(mise.NewProvider(), mise.NewMeta()) }
