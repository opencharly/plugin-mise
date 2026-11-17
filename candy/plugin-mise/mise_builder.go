package mise

// mise_builder.go — the builder:mise BUILD-TIME leg. The stage render lives in
// the SHARED sdk/kit.BuilderResolve (R3 — the ONE render the four detection
// builders already use); this package is only the composable selection point,
// exactly like candy/plugin-builder-pixi. The stage is DISTRO-AGNOSTIC: its
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

import (
	"encoding/json"
	"fmt"

	"github.com/opencharly/sdk/kit"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// builderWord is the reserved builder word this plugin serves.
const builderWord = "mise"

// invokeResolve handles the builder OpResolve: it decodes the host-computed
// BuilderResolveInput and dispatches to the shared kit.BuilderResolve (the ONE
// render, R3).
func invokeResolve(req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	var in spec.BuilderResolveInput
	if len(req.GetParamsJson()) > 0 {
		if err := json.Unmarshal(req.GetParamsJson(), &in); err != nil {
			return nil, fmt.Errorf("builder %q: decode resolve input: %w", builderWord, err)
		}
	}
	reply, err := kit.BuilderResolve(builderWord, in)
	if err != nil {
		return nil, err
	}
	j, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}
	return &pb.InvokeReply{ResultJson: j}, nil
}

// invokeCollectContext handles the builder OpCollectContext: the per-candy stage-context keys the
// host needs to drive this builder's multi-stage. It dispatches to the SAME shared
// kit.BuilderCollectContext the four plugin-builder-* plugins use, so mise's keys are the ones the
// shared render already produces rather than a second, drifting definition.
func invokeCollectContext(req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	var in spec.BuilderCollectInput
	if len(req.GetParamsJson()) > 0 {
		if err := json.Unmarshal(req.GetParamsJson(), &in); err != nil {
			return nil, fmt.Errorf("builder %q: decode collect-context input: %w", builderWord, err)
		}
	}
	j, err := json.Marshal(spec.BuilderCollectReply{Context: kit.BuilderCollectContext(builderWord, in)})
	if err != nil {
		return nil, err
	}
	return &pb.InvokeReply{ResultJson: j}, nil
}

// invokeReverse handles the builder OpReverse: the teardown ops for a resolved stage context. Same
// shared implementation as the collection leg, for the same reason.
func invokeReverse(req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	var in spec.BuilderReverseInput
	if len(req.GetParamsJson()) > 0 {
		if err := json.Unmarshal(req.GetParamsJson(), &in); err != nil {
			return nil, fmt.Errorf("builder %q: decode reverse input: %w", builderWord, err)
		}
	}
	j, err := json.Marshal(spec.BuilderReverseReply{ReverseOps: kit.BuilderReverse(builderWord, in)})
	if err != nil {
		return nil, err
	}
	return &pb.InvokeReply{ResultJson: j}, nil
}
