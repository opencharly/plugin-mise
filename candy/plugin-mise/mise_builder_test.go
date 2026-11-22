package mise

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/opencharly/sdk/kit"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// TestInvokeResolve_ReplyShape proves the OpResolve wire reply marshals a valid
// BuilderResolveReply rendered by the SHARED kit.BuilderResolve (R3) — the
// distro-agnostic stage driven by the host-computed input.
func TestInvokeResolve_ReplyShape(t *testing.T) {
	req := &pb.InvokeRequest{Op: "resolve", ParamsJson: []byte(`{"candy":"demo","builder_ref":"cachyos-builder","stage_name":"demo-mise-build","copy_src":"candy/demo","manifest":"mise.toml","uid":1000,"gid":1000,"home":"/home/user","install_cmd":"mise install"}`)}
	reply, err := provider{}.Invoke(context.Background(), req)
	if err != nil {
		t.Fatalf("Invoke(resolve): %v", err)
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
// The provider serves its OWN word on the deploy-time legs too: the collect-context and reverse
// replies equal kit's own for the literal word "mise", so a dispatch routed to any other builder
// fails here (opencharly/plugin-mise#26 review). An op no builder serves is still refused.
func TestInvoke_DispatchesTheDeployTimeBuilderOps(t *testing.T) {
	ctx := context.Background()

	collect := spec.BuilderCollectInput{Candy: "c", Builder: builderWord}
	params, _ := json.Marshal(collect)
	reply, err := provider{}.Invoke(ctx, &pb.InvokeRequest{Op: "collect-context", ParamsJson: params})
	if err != nil {
		t.Fatalf("Invoke(collect-context): %v", err)
	}
	wantCtx, err := kit.BuilderCollectContext("mise", collect)
	if err != nil {
		t.Fatalf("kit.BuilderCollectContext(mise): %v", err)
	}
	want, _ := json.Marshal(spec.BuilderCollectReply{Context: wantCtx, Package: kit.BuilderPackageOutput("mise", "c")})
	if !bytes.Equal(reply.GetResultJson(), want) {
		t.Errorf("collect-context served another builder's reply:\n got %s\nwant %s", reply.GetResultJson(), want)
	}

	rev := spec.BuilderReverseInput{Candy: "c", Builder: builderWord, Context: map[string]any{
		"tools": []any{"zig"}, "manifest": "mise.toml",
	}}
	params, _ = json.Marshal(rev)
	reply, err = provider{}.Invoke(ctx, &pb.InvokeRequest{Op: "reverse", ParamsJson: params})
	if err != nil {
		t.Fatalf("Invoke(reverse): %v", err)
	}
	ops := kit.BuilderReverse("mise", rev)
	if len(ops) == 0 {
		t.Fatalf("kit.BuilderReverse(mise) yielded no teardown ops for a full context; the pin would be vacuous")
	}
	wantRev, _ := json.Marshal(spec.BuilderReverseReply{ReverseOps: ops})
	if !bytes.Equal(reply.GetResultJson(), wantRev) {
		t.Errorf("reverse served another builder's teardown:\n got %s\nwant %s", reply.GetResultJson(), wantRev)
	}
}
