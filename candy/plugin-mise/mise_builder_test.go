package mise

import (
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
