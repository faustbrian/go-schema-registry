package avro_test

import (
	"context"
	"testing"

	schemaregistry "github.com/faustbrian/go-schema-registry"
	registryavro "github.com/faustbrian/go-schema-registry/formats/avro"
)

func FuzzAvroSchemas(f *testing.F) {
	f.Add([]byte(`"string"`))
	f.Add([]byte(`{"type":"record","name":"M","fields":[]}`))
	canonicalizer := registryavro.New(4096)
	f.Fuzz(func(t *testing.T, content []byte) {
		if len(content) == 0 || len(content) > 4096 {
			t.Skip()
		}
		_, _ = schemaregistry.Compile(context.Background(), schemaregistry.Definition{
			Format: schemaregistry.FormatAvro, Content: content,
		}, canonicalizer)
	})
}
