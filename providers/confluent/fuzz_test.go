package confluent_test

import (
	"context"
	"testing"

	"github.com/faustbrian/go-schema-registry/providers/confluent/v3"
)

// FuzzWireFrames probes at most 8192 frame bytes through both bounded version-0
// parsers; it does not validate descriptors, payload semantics or service behavior.
func FuzzWireFrames(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 1, 0})
	f.Add([]byte{1})
	// Signed64 values that previously aliased accepted counts/indexes on 32-bit.
	f.Add(signed64Frame(-(1 << 32)))
	f.Add(signed64Frame((1<<32)+1, 0))
	f.Add(signed64Frame(1, -(1 << 32)))
	f.Add(signed64Frame(1, (1<<32)+7))
	classic, err := confluent.NewClassicFramer("fuzz", 4096)
	if err != nil {
		f.Fatalf("NewClassicFramer() error = %v", err)
	}
	protobuf, err := confluent.NewProtobufFramer("fuzz", 4096, 32)
	if err != nil {
		f.Fatalf("NewProtobufFramer() error = %v", err)
	}
	f.Fuzz(func(t *testing.T, framed []byte) {
		if len(framed) > 8192 {
			t.Skip()
		}
		_, _, _ = classic.Unframe(context.Background(), framed)
		_, _, _, _ = protobuf.UnframeMessage(context.Background(), framed)
	})
}
