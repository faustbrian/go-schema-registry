package schemaregistry_test

import (
	"context"
	"testing"

	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
)

// BenchmarkCompilePortableIdentity measures copies/admission/SHA identity for a
// fixed small definition using an identity canonicalizer, not real Avro parsing.
func BenchmarkCompilePortableIdentity(b *testing.B) {
	definition := schemaregistry.Definition{Format: schemaregistry.FormatAvro, Content: []byte(`"string"`)}
	canonicalizer := canonicalizerFunc(func(_ context.Context, definition schemaregistry.Definition) ([]byte, error) {
		return definition.Content, nil
	})
	b.ReportAllocs()
	for b.Loop() {
		if _, err := schemaregistry.Compile(context.Background(), definition, canonicalizer); err != nil {
			b.Fatal(err)
		}
	}
}
