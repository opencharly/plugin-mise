package mise

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// TestRenderMiseShell_Install proves the `mise: install node@22` shorthand
// renders the install + reshim shell.
func TestRenderMiseShell_Install(t *testing.T) {
	shell, err := renderMiseShell(miseInput{Command: "install", Args: []string{"node@22"}})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if shell != "mise install node@22 && mise reshim" {
		t.Fatalf("unexpected shell: %q", shell)
	}
}

// TestRenderMiseShell_ToolShorthand proves the `tool:` shorthand.
func TestRenderMiseShell_ToolShorthand(t *testing.T) {
	shell, err := renderMiseShell(miseInput{Command: "use", Tool: "node@22"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if shell != "mise use node@22 && mise reshim" {
		t.Fatalf("unexpected shell: %q", shell)
	}
}

// TestRenderMiseShell_Task proves the `task:` shorthand.
func TestRenderMiseShell_Task(t *testing.T) {
	shell, err := renderMiseShell(miseInput{Command: "run", Task: "build"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if shell != "mise run build" {
		t.Fatalf("unexpected shell: %q", shell)
	}
}

// TestInvokeEmit_ActScript proves the OpEmit wire reply is an act-script
// EmitReply (the host wraps it in RUN).
func TestInvokeEmit_ActScript(t *testing.T) {
	req := &pb.InvokeRequest{ParamsJson: []byte(`{"plugin_input":{"command":"version"}}`)}
	reply, err := invokeEmit(context.Background(), req)
	if err != nil {
		t.Fatalf("invokeEmit: %v", err)
	}
	var er spec.EmitReply
	if err := json.Unmarshal(reply.GetResultJson(), &er); err != nil {
		t.Fatalf("reply is not an EmitReply: %v", err)
	}
	if !er.ActScript {
		t.Fatalf("expected act_script=true, got %+v", er)
	}
	if !strings.Contains(er.Fragment, "mise version") {
		t.Fatalf("fragment missing mise version: %q", er.Fragment)
	}
}
