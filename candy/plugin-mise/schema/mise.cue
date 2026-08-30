// Self-contained input schema for the builder:mise + verb:mise capabilities —
// references no base def, so it compiles standalone (the SDK's serve-side
// compile). The host validates every authored `mise:` step's plugin_input
// against this def; the plugin decodes the same shape into its Go input struct.
#MiseInput: {
	command: "install" | "use" | "run" | "exec" | "x" | "env" | "ls" | "doctor" | "version" | "shims" | "reshim" | "trust" | string
	args?: [...string]
	tool?: string
	task?: string
	config?: string
	tools?: { [string]: string }
	env?: { [string]: string }
	run_as?: string
}
