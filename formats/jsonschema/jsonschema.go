// Package jsonschema composes the registry /v3 contract with owned JSON Schema
// /v2 v2.0.0 for local validation and RFC 8785 JCS identity. New copies explicit
// resources; missing refs do not trigger remote discovery. Draft 2020-12 is the
// default, with explicit exported dialect choices. Adapter also supplies JSON
// business Encode/Decode; no registry, transport or shutdown lifecycle is owned.
// See SCHEMA-REG-DEC-002 and docs/migration-v3.md; schema validity/portable hashes
// are not authenticity, provider compatibility or every application policy.
package jsonschema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/deszhou/jcs"
	jsonschema "github.com/faustbrian/go-json-schema/v2"
	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
)

// ErrPayloadInvalid marks a value that does not satisfy the compiled JSON Schema.
var ErrPayloadInvalid = errors.New("schema registry json schema: invalid payload")

type diagnosticError struct {
	message string
	cause   error
}

// Error returns fixed category/operation text, not collaborator diagnostics.
func (err diagnosticError) Error() string { return err.message }

// Unwrap deliberately exposes retained causes for trusted inspection; those
// causes may contain sensitive text and are not safe diagnostic rendering.
func (err diagnosticError) Unwrap() error { return err.cause }

// Dialect is the explicit JSON Schema dialect selected for compilation.
type Dialect = jsonschema.Dialect

const (
	// Draft3 selects JSON Schema draft 3.
	Draft3 = jsonschema.Draft3
	// Draft4 selects JSON Schema draft 4.
	Draft4 = jsonschema.Draft4
	// Draft6 selects JSON Schema draft 6.
	Draft6 = jsonschema.Draft6
	// Draft7 selects JSON Schema draft 7.
	Draft7 = jsonschema.Draft7
	// Draft201909 selects JSON Schema draft 2019-09.
	Draft201909 = jsonschema.Draft201909
	// Draft202012 selects JSON Schema draft 2020-12.
	Draft202012 = jsonschema.Draft202012
)

// Config bounds schema compilation and payload validation.
type Config struct {
	// Dialect is zero for Draft202012, otherwise an explicit exported dialect.
	Dialect Dialect
	// Resources is an explicit URI-to-document map, cloned and byte-copied.
	// Only these local documents are returned by the resource loader.
	Resources map[string][]byte
	// MaxSchemaBytes is positive and caps root/resource text, not all parser memory.
	MaxSchemaBytes int
	// MaxTotalSchemaBytes bounds copied resource sum and compiler total schema
	// bytes; caller must use finite supported limits, not unchecked machine maxima.
	MaxTotalSchemaBytes int
	// MaxPayloadBytes is positive; Encode checks it after json.Marshal allocation.
	MaxPayloadBytes int
	// MaxResources is positive and caps explicit resource entries; compiler
	// admission adds one for the root. Other compiler limits retain owned defaults.
	MaxResources int
}

// Adapter validates JSON Schema documents and JSON payloads without implicit
// resource loading.
type Adapter struct {
	compiler        *jsonschema.Compiler
	maxSchemaBytes  int
	maxPayloadBytes int
}

// New constructs an adapter using Draft 2020-12 by default and no remote loader.
// It validates positive limits and copies resource bytes, while borrowing no
// application endpoint or registry. Shared reuse follows the owned compiler's
// concurrency contract. Use New; nil/zero-value adapters are not usable.
func New(config Config) (*Adapter, error) {
	if config.Dialect == "" {
		config.Dialect = Draft202012
	}
	if config.MaxSchemaBytes <= 0 || config.MaxTotalSchemaBytes <= 0 ||
		config.MaxPayloadBytes <= 0 || config.MaxResources <= 0 || len(config.Resources) > config.MaxResources {
		return nil, fmt.Errorf("json schema adapter limits must be positive")
	}
	resources := maps.Clone(config.Resources)
	totalBytes := 0
	for identifier, resource := range resources {
		if identifier == "" || len(resource) == 0 || len(resource) > config.MaxSchemaBytes {
			return nil, fmt.Errorf("invalid JSON Schema resource")
		}
		totalBytes += len(resource)
		if totalBytes > config.MaxTotalSchemaBytes {
			return nil, fmt.Errorf("JSON Schema resources exceed total byte limit")
		}
		resources[identifier] = append([]byte(nil), resource...)
	}
	limits := jsonschema.DefaultLimits()
	limits.MaxInputBytes = max(config.MaxSchemaBytes, config.MaxPayloadBytes)
	limits.MaxTotalSchemaBytes = config.MaxTotalSchemaBytes
	limits.MaxSchemaResources = config.MaxResources + 1
	options := []jsonschema.Option{jsonschema.WithDialect(config.Dialect), jsonschema.WithLimits(limits)}
	if len(resources) != 0 {
		options = append(options, jsonschema.WithResourceLoader(jsonschema.ResourceLoaderFunc(
			func(ctx context.Context, identifier string) ([]byte, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				resource, found := resources[identifier]
				if !found {
					return nil, jsonschema.ErrResourceNotFound
				}
				return append([]byte(nil), resource...), nil
			},
		)))
	}
	compiler, err := jsonschema.NewCompiler(options...)
	if err != nil {
		return nil, diagnosticError{"construct JSON Schema compiler failed", err}
	}
	return &Adapter{
		compiler:        compiler,
		maxSchemaBytes:  config.MaxSchemaBytes,
		maxPayloadBytes: config.MaxPayloadBytes,
	}, nil
}

// Canonicalize validates with golib/json-schema and applies RFC 8785 JCS. JCS
// deliberately restricts portable fingerprints to I-JSON number semantics.
// It reads borrowed Content and returns owned canonical bytes; format/reference
// validation uses only configured resources, not provider reference coordinates.
// JCS follows successful compiler validation and is synchronous without ctx;
// core Compile is responsible for canonical-output admission afterward.
func (adapter *Adapter) Canonicalize(
	ctx context.Context,
	definition schemaregistry.Definition,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if definition.Format != schemaregistry.FormatJSONSchema {
		return nil, fmt.Errorf("unsupported schema format")
	}
	if len(definition.Content) > adapter.maxSchemaBytes {
		return nil, fmt.Errorf("schema exceeds %d bytes", adapter.maxSchemaBytes)
	}
	if _, err := adapter.compiler.Compile(ctx, definition.Content); err != nil {
		return nil, diagnosticError{"compile JSON Schema failed", err}
	}
	canonical, err := jcs.Transform(definition.Content)
	if err != nil {
		return nil, diagnosticError{"canonicalize JSON Schema failed", err}
	}
	return canonical, nil
}

// Encode marshals the caller value, then caps/validates resulting JSON against
// the supplied schema. json.Marshaler callbacks may allocate/block/panic before
// the byte check and are not preempted. The target schema is recompiled locally;
// there is no provider lookup or promise of a total allocation bound.
func (adapter *Adapter) Encode(ctx context.Context, schema schemaregistry.Schema, value any) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, diagnosticError{"marshal JSON payload failed", err}
	}
	if len(payload) > adapter.maxPayloadBytes {
		return nil, fmt.Errorf("%w: payload exceeds %d bytes", schemaregistry.ErrLimitExceeded, adapter.maxPayloadBytes)
	}
	if err := adapter.validate(ctx, schema, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// Decode caps and validates borrowed payload before json.Unmarshal into target.
// On validation failure target is untouched by this adapter; custom unmarshal
// callbacks or conversion errors can partially mutate it afterward. Those calls
// are not context-preempted or panic-contained. Retained causes are trusted
// inspection data, not safe log text; no secure erasure is supplied.
func (adapter *Adapter) Decode(
	ctx context.Context,
	schema schemaregistry.Schema,
	payload []byte,
	target any,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(payload) > adapter.maxPayloadBytes {
		return fmt.Errorf("%w: payload exceeds %d bytes", schemaregistry.ErrLimitExceeded, adapter.maxPayloadBytes)
	}
	if err := adapter.validate(ctx, schema, payload); err != nil {
		return err
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return diagnosticError{"unmarshal JSON payload failed", err}
	}
	return nil
}

func (adapter *Adapter) validate(ctx context.Context, schema schemaregistry.Schema, payload []byte) error {
	definition := schema.Definition()
	if definition.Format != schemaregistry.FormatJSONSchema || len(definition.Content) > adapter.maxSchemaBytes {
		return fmt.Errorf("%w: schema format or size", schemaregistry.ErrInvalidSchema)
	}
	compiled, err := adapter.compiler.Compile(ctx, definition.Content)
	if err != nil {
		return diagnosticError{"schema registry: invalid schema: compilation failed", schemaregistry.ErrInvalidSchema}
	}
	result, err := compiled.Validate(ctx, payload)
	if err != nil {
		return diagnosticError{"schema registry json schema: invalid payload", ErrPayloadInvalid}
	}
	if !result.Valid {
		return ErrPayloadInvalid
	}
	return nil
}
