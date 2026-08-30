package mise

// mise_verb.go — the verb:mise legs. OpEmit renders the shell a `mise:` plan
// step runs at image build (returned as an act-script — the host wraps it in a
// RUN with the step's run_as); OpExecute runs `mise <command>` on the venue at
// check/deploy time via the check-engine reverse channel.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opencharly/sdk"
	"github.com/opencharly/sdk/kit"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// miseInput is the plugin-side decode of the authored `mise:` step input
// (schema/mise.cue #MiseInput). The host validates the authored input against
// the served CUE schema; this struct is the Go decode for dispatch.
type miseInput struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Tool    string            `json:"tool"`
	Task    string            `json:"task"`
	Config  string            `json:"config"`
	Tools   map[string]string `json:"tools"`
	Env     map[string]string `json:"env"`
	RunAs   string            `json:"run_as"`
}

// decodeInput decodes the authored plugin_input (the desugared `mise:` step
// value) from the InvokeRequest params. The params carry the full #Op; the
// plugin_input is the per-verb field set (the schema-compaction cutover shape).
func decodeInput(req *pb.InvokeRequest) (miseInput, error) {
	var in miseInput
	if len(req.GetParamsJson()) == 0 {
		return in, nil
	}
	var op spec.Op
	if err := json.Unmarshal(req.GetParamsJson(), &op); err != nil {
		return in, fmt.Errorf("decode op: %w", err)
	}
	kit.DecodeInput(op.PluginInput, &in)
	return in, nil
}

// renderMiseShell renders the shell a `mise:` step runs. The command defaults to
// `install`; `tool`/`task` are shorthands for `mise use/install <tool>` and
// `mise run <task>`. Installs are followed by `mise reshim` so the image's shims
// exist for the installed tools.
func renderMiseShell(in miseInput) (string, error) {
	cmd := in.Command
	if cmd == "" {
		cmd = "install"
	}
	var parts []string
	switch {
	case in.Tool != "" && (cmd == "use" || cmd == "install"):
		parts = append(parts, "mise "+cmd+" "+in.Tool)
	case in.Task != "" && cmd == "run":
		parts = append(parts, "mise run "+in.Task)
	default:
		parts = append(parts, "mise "+cmd+" "+strings.Join(in.Args, " "))
	}
	if cmd == "install" || cmd == "use" {
		parts = append(parts, "mise reshim")
	}
	return strings.TrimSpace(strings.Join(parts, " && ")), nil
}

// invokeEmit handles the verb OpEmit at image build: it renders the shell and
// returns it as an act-script (the host wraps it in a RUN with the step's
// run_as).
func invokeEmit(_ context.Context, req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	in, err := decodeInput(req)
	if err != nil {
		return nil, err
	}
	shell, err := renderMiseShell(in)
	if err != nil {
		return nil, err
	}
	j, err := json.Marshal(spec.EmitReply{Fragment: shell, ActScript: true})
	if err != nil {
		return nil, err
	}
	return &pb.InvokeReply{ResultJson: j}, nil
}

// invokeExecute handles the verb OpExecute at check/deploy time: it runs
// `mise <command>` on the venue via the check-engine reverse channel and
// returns the wire verdict.
func invokeExecute(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	in, err := decodeInput(req)
	if err != nil {
		return sdk.ResultJSON("fail", "mise: "+err.Error())
	}
	shell, err := renderMiseShell(in)
	if err != nil {
		return sdk.ResultJSON("fail", "mise: "+err.Error())
	}
	cc, err := sdk.NewCheckContext(req.GetExecutorBrokerId(), req.GetEnvJson())
	if err != nil {
		return sdk.ResultJSON("fail", "mise: reverse channel: "+err.Error())
	}
	stdout, stderr, exit, runErr := cc.Exec().RunCapture(ctx, shell)
	if runErr != nil {
		return sdk.ResultJSON("fail", "mise: "+runErr.Error())
	}
	if exit != 0 {
		return sdk.ResultJSON("fail", fmt.Sprintf("mise %s: exit %d: %s", in.Command, exit, stderr))
	}
	return sdk.ResultJSON("pass", stdout)
}
