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
	case sdk.OpEmit:
		return invokeEmit(ctx, req)
	case sdk.OpExecute:
		return invokeExecute(ctx, req)
	}
	return nil, fmt.Errorf("mise: unsupported op %q (serves only %q, %q, %q)", req.GetOp(), sdk.OpResolve, sdk.OpEmit, sdk.OpExecute)
}
