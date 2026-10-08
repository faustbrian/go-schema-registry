package jsonschema_test

import (
	"errors"
	"testing"

	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
	registryjsonschema "github.com/faustbrian/go-schema-registry/v3/formats/jsonschema"
)

func TestRegistryJSONSchemaStandardUnicodeProperty(t *testing.T) {
	t.Parallel()
	adapter, err := registryjsonschema.New(registryjsonschema.Config{
		MaxSchemaBytes: 4096, MaxTotalSchemaBytes: 8192, MaxPayloadBytes: 64, MaxResources: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := schemaregistry.Compile(t.Context(), schemaregistry.Definition{
		Format:  schemaregistry.FormatJSONSchema,
		Content: []byte(`{"type":"string","pattern":"^\\p{Script=Greek}+$"}`),
	}, adapter)
	if err != nil {
		t.Fatalf("compile standard Unicode property: %v", err)
	}
	encoded, err := adapter.Encode(t.Context(), schema, "α")
	if err != nil || string(encoded) != `"α"` {
		t.Fatalf("encode Greek value: error=%v", err)
	}
	var decoded string
	if err := adapter.Decode(t.Context(), schema, encoded, &decoded); err != nil || decoded != "α" {
		t.Fatalf("decode Greek value: error=%v", err)
	}
	if _, err := adapter.Encode(t.Context(), schema, "A"); !errors.Is(err, registryjsonschema.ErrPayloadInvalid) {
		t.Fatalf("Latin value: got %v, want ErrPayloadInvalid", err)
	}
	decoded = "unchanged"
	if err := adapter.Decode(t.Context(), schema, []byte(`"A"`), &decoded); !errors.Is(err, registryjsonschema.ErrPayloadInvalid) || decoded != "unchanged" {
		t.Fatalf("invalid decode mutated target or lost error: error=%v", err)
	}
}

func TestRegistryJSONSchemaURITemplateAssertionIsExplicit(t *testing.T) {
	t.Parallel()
	const dialect = "https://schemas.example/format-assertion"
	meta := []byte(`{"$schema":"https://json-schema.org/draft/2020-12/schema","$vocabulary":{"https://json-schema.org/draft/2020-12/vocab/core":true,"https://json-schema.org/draft/2020-12/vocab/validation":true,"https://json-schema.org/draft/2020-12/vocab/format-assertion":true}}`)
	for _, asserted := range []bool{false, true} {
		name := "annotation"
		content := []byte(`{"type":"string","format":"uri-template"}`)
		config := registryjsonschema.Config{
			MaxSchemaBytes: 4096, MaxTotalSchemaBytes: 8192, MaxPayloadBytes: 64, MaxResources: 2,
		}
		if asserted {
			name = "assertion"
			content = []byte(`{"$schema":"` + dialect + `","type":"string","format":"uri-template"}`)
			config.Resources = map[string][]byte{dialect: meta}
		}
		t.Run(name, func(t *testing.T) {
			adapter, err := registryjsonschema.New(config)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := schemaregistry.Compile(t.Context(), schemaregistry.Definition{
				Format: schemaregistry.FormatJSONSchema, Content: content,
			}, adapter)
			if err != nil {
				t.Fatalf("compile URI-template policy: %v", err)
			}
			for _, test := range []struct {
				value string
				valid bool
			}{{"{var:1}", true}, {"{var*}", true}, {"{var:1*}", false}, {"{var:01}", false}} {
				encoded, err := adapter.Encode(t.Context(), schema, test.value)
				if test.valid || !asserted {
					if err != nil {
						t.Fatalf("accepted URI-template: %v", err)
					}
					var decoded string
					if err := adapter.Decode(t.Context(), schema, encoded, &decoded); err != nil || decoded != test.value {
						t.Fatalf("URI-template round trip: %v", err)
					}
				} else if !errors.Is(err, registryjsonschema.ErrPayloadInvalid) || encoded != nil {
					t.Fatalf("invalid asserted URI-template: got %v, want ErrPayloadInvalid", err)
				}
			}
		})
	}
}
