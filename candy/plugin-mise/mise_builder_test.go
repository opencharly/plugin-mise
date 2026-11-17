package mise

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// TestInvokeResolve_ReplyShape proves the OpResolve wire reply marshals a valid
// BuilderResolveReply rendered by the SHARED kit.BuilderResolve (R3) — the
// distro-agnostic stage driven by the host-computed input.
func TestInvokeResolve_ReplyShape(t *testing.T) {
	req := &pb.InvokeRequest{ParamsJson: []byte(`{"candy":"demo","builder_ref":"cachyos-builder","stage_name":"demo-mise-build","copy_src":"candy/demo","manifest":"mise.toml","uid":1000,"gid":1000,"home":"/home/user","install_cmd":"mise install"}`)}
	reply, err := invokeResolve(req)
	if err != nil {
		t.Fatalf("invokeResolve: %v", err)
	}
	var brr spec.BuilderResolveReply
	if err := json.Unmarshal(reply.GetResultJson(), &brr); err != nil {
		t.Fatalf("reply is not a BuilderResolveReply: %v", err)
	}
	if !strings.Contains(brr.Stage, "FROM cachyos-builder AS demo-mise-build") {
		t.Fatalf("stage missing BuilderRef FROM: %q", brr.Stage)
	}
	if !strings.Contains(brr.Stage, "COPY --chown=1000:1000 candy/demo/mise.toml mise.toml") {
		t.Fatalf("stage missing manifest COPY: %q", brr.Stage)
	}
	if !strings.Contains(brr.Stage, "mise install && mise reshim") {
		t.Fatalf("stage missing install+reshim: %q", brr.Stage)
	}
	if len(brr.CopyArtifacts) != 1 || !strings.Contains(brr.CopyArtifacts[0], "/home/user") {
		t.Fatalf("artifacts must copy the Home: %v", brr.CopyArtifacts)
	}
	if !strings.Contains(brr.CopyBinary, "/usr/local/bin/mise") {
		t.Fatalf("binary copy missing: %q", brr.CopyBinary)
	}
}

// TestInvoke_DispatchesTheDeployTimeBuilderOps is the coverage the bed cannot
// give: `Invoke` must answer BOTH deploy-time builder legs for a mise-composed
// image, and only collect-context is reachable from a bed's deploy path in this
// repo (no `disposable: true` target exercises teardown). Without the two
// dispatch cases this test fails with the production error
//
//	mise: unsupported op "collect-context"
//
// which is exactly the defect opencharly/plugin-mise#22 reported.
func TestInvoke_DispatchesTheDeployTimeBuilderOps(t *testing.T) {
	for _, op := range []string{"collect-context", "reverse"} {
		reply, err := provider{}.Invoke(context.Background(), &pb.InvokeRequest{Op: op})
		if err != nil {
			t.Fatalf("Invoke(%q): %v", op, err)
		}
		if len(reply.GetResultJson()) == 0 {
			t.Fatalf("Invoke(%q): empty reply", op)
		}
		var probe map[string]any
		if err := json.Unmarshal(reply.GetResultJson(), &probe); err != nil {
			t.Fatalf("Invoke(%q): reply is not JSON: %v", op, err)
		}
	}
}
