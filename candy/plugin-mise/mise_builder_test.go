package mise

import (
	"encoding/json"
	"strings"
	"testing"

	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// TestRenderMiseStage_Minimal proves the external_builder path (minimal input)
// renders a deterministic, distro-aware stage that provisions mise + the system
// config, with the CopyArtifacts + InlineFragment the host splices.
func TestRenderMiseStage_Minimal(t *testing.T) {
	reply := renderMiseStage(spec.BuilderResolveInput{Candy: "demo"}, spec.BuildEnv{Distros: []string{"fedora"}})
	if !strings.Contains(reply.Stage, "FROM quay.io/fedora/fedora:43 AS mise-build") {
		t.Fatalf("stage missing distro base: %q", reply.Stage)
	}
	if !strings.Contains(reply.Stage, "mise-"+miseVersion+"-linux-$A.tar.gz") {
		t.Fatalf("stage missing pinned mise tarball: %q", reply.Stage)
	}
	if !strings.Contains(reply.Stage, "strip-components=1") {
		t.Fatalf("stage missing strip-components: %q", reply.Stage)
	}
	if len(reply.CopyArtifacts) != 3 {
		t.Fatalf("expected 3 copy artifacts, got %d: %v", len(reply.CopyArtifacts), reply.CopyArtifacts)
	}
	if !strings.Contains(reply.CopyArtifacts[2], "/etc/mise/config.toml") {
		t.Fatalf("config artifact missing: %v", reply.CopyArtifacts)
	}
	if !strings.Contains(reply.InlineFragment, "MISE_DATA_DIR=/usr/local/share/mise") {
		t.Fatalf("inline fragment missing MISE_DATA_DIR: %q", reply.InlineFragment)
	}
	if !strings.Contains(reply.InlineFragment, "shims:$PATH") {
		t.Fatalf("inline fragment missing shims PATH: %q", reply.InlineFragment)
	}
}

// TestRenderMiseStage_Full proves the P3 detection path (full input) copies the
// candy's config into the stage and runs `mise install` there.
func TestRenderMiseStage_Full(t *testing.T) {
	reply := renderMiseStage(spec.BuilderResolveInput{
		Candy:            "demo",
		CopySrc:          "candy/demo",
		Manifest:         "mise.toml",
		UID:              1000,
		GID:              1000,
		CacheMountsOwned: "--mount=type=cache,target=/mise ",
	}, spec.BuildEnv{Distros: []string{"debian"}})
	if !strings.Contains(reply.Stage, "COPY --chown=1000:1000 candy/demo/mise.toml mise.toml") {
		t.Fatalf("full stage missing config COPY: %q", reply.Stage)
	}
	if !strings.Contains(reply.Stage, "mise install && mise reshim") {
		t.Fatalf("full stage missing mise install: %q", reply.Stage)
	}
	if !strings.Contains(reply.Stage, "FROM debian:13 AS mise-build") {
		t.Fatalf("full stage missing debian base: %q", reply.Stage)
	}
}

// TestInvokeResolve_ReplyShape proves the OpResolve wire reply marshals a valid
// BuilderResolveReply (the shape the host splices).
func TestInvokeResolve_ReplyShape(t *testing.T) {
	req := &pb.InvokeRequest{ParamsJson: []byte(`{"candy":"demo"}`), EnvJson: []byte(`{"distros":["fedora"]}`)}
	reply, err := invokeResolve(req)
	if err != nil {
		t.Fatalf("invokeResolve: %v", err)
	}
	var brr spec.BuilderResolveReply
	if err := json.Unmarshal(reply.GetResultJson(), &brr); err != nil {
		t.Fatalf("reply is not a BuilderResolveReply: %v", err)
	}
	if !strings.Contains(brr.Stage, "AS mise-build") {
		t.Fatalf("reply stage missing: %q", brr.Stage)
	}
}
