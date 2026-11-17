package mise

import (
	"context"
	"fmt"

	"github.com/opencharly/sdk"
	pb "github.com/opencharly/spec/proto"
)

// provider is the mise provider: the ONE Invoke dispatch surface for both the
// builder and verb capabilities, keyed by op.
type provider struct{ pb.UnimplementedProviderServer }

// Invoke dispatches the build-time builder leg (OpResolve → BuilderResolveReply),
// the build-time verb leg (OpEmit → EmitReply) and the check/deploy-time verb
// leg (OpExecute → verdict). Any other op is a loud error.
func (provider) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	switch req.GetOp() {
	case sdk.OpResolve:
		return invokeResolve(req)
	// OpCollectContext and OpReverse are the DEPLOY-TIME legs of an externalized
	// detection-builder, and the host asks EVERY builder for them. The four
	// plugin-builder-* plugins implement both; this one did not, so any bed that
	// deploys a mise-composed image died with
	//   builder "mise" collect-context: ... desc = mise: unsupported op
	// which reads like a plugin defect and was one. opencharly/plugin-mise#22.
	case sdk.OpCollectContext:
		return invokeCollectContext(req)
	case sdk.OpReverse:
		return invokeReverse(req)
	case sdk.OpEmit:
		return invokeEmit(ctx, req)
	case sdk.OpExecute:
		return invokeExecute(ctx, req)
	}
	return nil, fmt.Errorf("mise: unsupported op %q (serves only %q, %q, %q, %q, %q)", req.GetOp(), sdk.OpResolve, sdk.OpCollectContext, sdk.OpReverse, sdk.OpEmit, sdk.OpExecute)
}
