package glue_test

import (
	"context"
	"testing"

	registryglue "github.com/faustbrian/go-schema-registry/providers/glue/v2"
)

// FuzzWireFrames probes bounded header-version-3 parsing, including unsupported
// compression; it does not exercise the SDK or authenticate UUID/payload content.
func FuzzWireFrames(f *testing.F) {
	f.Add([]byte{3, 0, 0x12, 0x3e, 0x45, 0x67, 0xe8, 0x9b, 0x12, 0xd3, 0xa4, 0x56, 0x42, 0x66, 0x14, 0x17, 0x40, 0x00})
	f.Add([]byte{3, 5})
	framer, err := registryglue.NewUncompressedFramer("fuzz", 4096)
	if err != nil {
		f.Fatalf("NewUncompressedFramer() error = %v", err)
	}
	f.Fuzz(func(t *testing.T, framed []byte) {
		if len(framed) > 8192 {
			t.Skip()
		}
		_, _, _ = framer.Unframe(context.Background(), framed)
	})
}
