package avro_test

import (
	"context"
	"strings"
	"testing"

	registry "github.com/faustbrian/go-schema-registry"
	"github.com/faustbrian/go-schema-registry/formats/avro"
)

func TestPrivacyCanonicalizerDiagnostic(t *testing.T) {
	_, err := avro.New(1024).Canonicalize(context.Background(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`"application-private-detail"`)})
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Fatalf("private Avro diagnostic: %v", err)
	}
}
