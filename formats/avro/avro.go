// Package avro validates standalone Avro and returns parsing canonical form
// using goavro/v2 v2.15.0 against the pinned Avro 1.12.0 profile. New supplies
// an immutable size policy; no registry, remote named-schema loader or value
// codec is owned. Inline external definitions explicitly before Canonicalize.
// See SCHEMA-REG-DEC-001 and spec/sources.lock.json; this is not authentication
// or a claim of arbitrary provider compatibility.
package avro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
	"github.com/linkedin/goavro/v2"
)

const maxNestingDepth = 256

type diagnosticError struct {
	message string
	cause   error
}

// Error returns fixed category/operation text, not collaborator diagnostics.
func (err diagnosticError) Error() string { return err.message }

// Unwrap deliberately exposes retained causes for trusted inspection; those
// causes may contain sensitive text and are not safe diagnostic rendering.
func (err diagnosticError) Unwrap() error { return err.cause }

// Canonicalizer validates and canonicalizes bounded standalone Avro schemas.
type Canonicalizer struct{ maxSchemaBytes int }

// New constructs an Avro canonicalizer. External named-schema references must
// be resolved and inlined explicitly before this focused adapter is called.
// Construction does not validate maxSchemaBytes: nonpositive values, zero-value
// or nil canonicalizers fail when Canonicalize is invoked. No resources need Close.
func New(maxSchemaBytes int) *Canonicalizer {
	return &Canonicalizer{maxSchemaBytes: maxSchemaBytes}
}

// Canonicalize rejects wrong format, oversized text and nesting above 256 before
// invoking goavro. It reads borrowed Content synchronously and ignores provider
// coordinates/metadata; reference bodies must already be inline. Context is
// checked between JSON tokens and after codec construction, not inside goavro.
// Returned bytes are owned; parsing canonical form omits properties such as docs
// and defaults, so equal identities are not identical source text/application policy.
// Error() is categorical; explicit cause inspection can expose parser diagnostics.
func (canonicalizer *Canonicalizer) Canonicalize(
	ctx context.Context,
	definition schemaregistry.Definition,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if canonicalizer == nil || canonicalizer.maxSchemaBytes <= 0 {
		return nil, fmt.Errorf("invalid Avro canonicalizer")
	}
	if definition.Format != schemaregistry.FormatAvro {
		return nil, fmt.Errorf("unsupported schema format")
	}
	if len(definition.Content) > canonicalizer.maxSchemaBytes {
		return nil, fmt.Errorf("schema exceeds %d bytes", canonicalizer.maxSchemaBytes)
	}
	if err := checkNesting(ctx, definition.Content); err != nil {
		return nil, err
	}
	codec, err := goavro.NewCodec(string(definition.Content))
	if err != nil {
		return nil, diagnosticError{"parse Avro schema failed", err}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []byte(codec.CanonicalSchema()), nil
}

// Stream JSON tokens to reject deep nesting before the recursive Avro parser;
// this is a depth/byte admission guard, not a peak-allocation or hard-time cap.
func checkNesting(ctx context.Context, content []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	depth := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return diagnosticError{"parse Avro schema JSON failed", err}
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			continue
		}
		switch delimiter {
		case '{', '[':
			depth++
			if depth > maxNestingDepth {
				return fmt.Errorf("avro schema nesting exceeds %d", maxNestingDepth)
			}
		case '}', ']':
			depth--
		}
	}
}
