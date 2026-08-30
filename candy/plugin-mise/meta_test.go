package mise

import (
	"context"
	"strings"
	"testing"
)

// TestNewMeta_SchemaSplices proves the served CUE schema compiles standalone and
// the capabilities advertise the builder + verb words (the schema-over-Describe
// contract the host validates authored inputs against).
func TestNewMeta_SchemaSplices(t *testing.T) {
	meta := NewMeta()
	caps, err := meta.Describe(context.Background(), nil)
	if err != nil {
		t.Fatalf("Describe: %v", err)
	}
	if caps == nil {
		t.Fatal("nil capabilities")
	}
	if !strings.Contains(caps.GetSchemaCue(), "#MiseInput") {
		t.Fatalf("schema missing #MiseInput: %q", caps.GetSchemaCue())
	}
	found := map[string]bool{}
	for _, c := range caps.GetProvided() {
		found[c.GetClass()+"."+c.GetWord()] = true
	}
	if !found["builder.mise"] || !found["verb.mise"] {
		t.Fatalf("capabilities missing builder.mise/verb.mise: %v", found)
	}
}
