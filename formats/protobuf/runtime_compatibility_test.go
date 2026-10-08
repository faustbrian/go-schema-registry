package protobuf_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	registry "github.com/faustbrian/go-schema-registry/v3"
	pb "github.com/faustbrian/go-schema-registry/v3/formats/protobuf"
)

// Prior runtime output is a literal compatibility oracle, not a second
// compilation with the currently selected protobuf dependency.
func TestCanonicalizerPreservesPriorRuntimeIdentities(t *testing.T) {
	t.Parallel()
	encoded, err := os.ReadFile("testdata/runtime-1.36.11-identities.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name        string
			Canonical   []byte
			Fingerprint string
			Bundle      []byte
		}
	}
	if err := json.Unmarshal(encoded, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 5 {
		t.Fatal("incomplete prior-runtime fixtures")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			var imports map[string]string
			if row.Name == "imports_options" {
				imports = map[string]string{"shared.proto": `syntax="proto3";package example;message Shared{string value=1;}`}
			}
			canonicalizer, err := pb.New(pb.Config{Filename: "event.proto", Imports: imports, MaxSchemaBytes: 8192, MaxImports: 4})
			if err != nil {
				t.Fatal(err)
			}
			bundle, err := registry.LoadBundle(context.Background(), row.Bundle, map[registry.Format]registry.Canonicalizer{registry.FormatProtobuf: canonicalizer}, registry.GraphLimits{MaxSchemas: 4, MaxDepth: 4, MaxReferences: 4}, 65536)
			if err != nil {
				t.Fatalf("prior bundle no longer loads: %v", err)
			}
			expected, err := registry.ParseFingerprint(row.Fingerprint)
			if err != nil {
				t.Fatal(err)
			}
			schema, ok := bundle.Resolve(expected)
			if !ok || schema.Fingerprint() != expected || bundle.Root().Fingerprint() != expected || !bytes.Equal(schema.Canonical(), row.Canonical) {
				t.Fatal("prior canonical identity changed")
			}
		})
	}
}
