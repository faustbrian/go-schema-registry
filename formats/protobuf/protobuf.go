// Package protobuf compiles one explicit local Protocol Buffers root with
// protocompile v0.14.1 and deterministically marshals its descriptor-set profile
// with protobuf runtime v1.36.12. New copies configured imports; there is no
// filesystem/network discovery or registry/value-codec lifecycle. The pinned
// language authority is v33.4; SCHEMA-REG-DEC-003 and the saved prior-runtime
// identities distinguish package identity from provider equivalence.
package protobuf

import (
	"context"
	"fmt"
	"maps"
	"math/bits"

	"github.com/bufbuild/protocompile"
	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Config provides an explicit root filename and immutable local import set.
type Config struct {
	// Filename is a required logical root path, not an automatically opened file.
	Filename string
	// Imports maps logical filenames to source text; copied at New and used
	// only by the map-backed resolver. Root filename may not appear here.
	Imports map[string]string
	// MaxSchemaBytes caps root text and independently aggregate import names/text.
	// The import map is cloned before aggregate text admission, not a total heap cap.
	MaxSchemaBytes int
	// MaxImports is a positive configured import count limit.
	MaxImports int
}

type diagnosticError struct {
	message string
	cause   error
}

// Error returns fixed category/operation text, not collaborator diagnostics.
func (err diagnosticError) Error() string { return err.message }

// Unwrap deliberately exposes retained causes for trusted inspection; those
// causes may contain sensitive text and are not safe diagnostic rendering.
func (err diagnosticError) Unwrap() error { return err.cause }

// Canonicalizer owns a copied local source set and performs no network access.
type Canonicalizer struct {
	filename       string
	imports        map[string]string
	maxSchemaBytes int
}

// New validates and copies Protobuf source configuration.
func New(config Config) (*Canonicalizer, error) {
	if config.Filename == "" || config.MaxSchemaBytes <= 0 || config.MaxImports <= 0 {
		return nil, fmt.Errorf("invalid Protobuf canonicalizer config")
	}
	if len(config.Imports) > config.MaxImports {
		return nil, fmt.Errorf("imports exceed %d", config.MaxImports)
	}
	imports := maps.Clone(config.Imports)
	if imports == nil {
		imports = make(map[string]string)
	}
	importBytes := uint64(0)
	for filename, source := range imports {
		if filename == "" || filename == config.Filename {
			return nil, fmt.Errorf("invalid Protobuf import")
		}
		var withinLimit bool
		importBytes, withinLimit = boundedTextBytes(importBytes, len(filename), config.MaxSchemaBytes)
		if !withinLimit {
			return nil, fmt.Errorf("protobuf imports exceed %d bytes", config.MaxSchemaBytes)
		}
		importBytes, withinLimit = boundedTextBytes(importBytes, len(source), config.MaxSchemaBytes)
		if !withinLimit {
			return nil, fmt.Errorf("protobuf imports exceed %d bytes", config.MaxSchemaBytes)
		}
	}
	return &Canonicalizer{filename: config.Filename, imports: imports, maxSchemaBytes: config.MaxSchemaBytes}, nil
}

func boundedTextBytes(current uint64, additional, limit int) (uint64, bool) {
	// #nosec G115 -- New supplies nonnegative string lengths after validating its positive MaxSchemaBytes limit.
	total, carry := bits.Add64(current, uint64(additional), 0)
	// #nosec G115 -- New rejects nonpositive MaxSchemaBytes before these sums; carry rejects overflow.
	return total, carry == 0 && total <= uint64(limit)
}

// Canonicalize compiles the root and explicit imports, then deterministically
// marshals the returned root descriptors; linked import text is supplied only
// through the copied local map, not fetched from provider Reference coordinates.
// Callers must bind portable dependency fingerprints consistently themselves.
// It borrows Content during compilation, returns owned bytes, and uses one
// compiler worker. Use New; nil/zero-value receivers are not supported.
// No postmarshal output cap is applied here; core Compile adds its own result cap.
func (canonicalizer *Canonicalizer) Canonicalize(
	ctx context.Context,
	definition schemaregistry.Definition,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if definition.Format != schemaregistry.FormatProtobuf {
		return nil, fmt.Errorf("unsupported schema format")
	}
	if len(definition.Content) > canonicalizer.maxSchemaBytes {
		return nil, fmt.Errorf("schema exceeds %d bytes", canonicalizer.maxSchemaBytes)
	}
	sources := maps.Clone(canonicalizer.imports)
	sources[canonicalizer.filename] = string(definition.Content)
	compiler := protocompile.Compiler{
		Resolver:       &protocompile.SourceResolver{Accessor: protocompile.SourceAccessorFromMap(sources)},
		MaxParallelism: 1,
	}
	files, err := compiler.Compile(ctx, canonicalizer.filename)
	if err != nil {
		return nil, diagnosticError{"compile Protobuf schema failed", err}
	}
	set := &descriptorpb.FileDescriptorSet{File: make([]*descriptorpb.FileDescriptorProto, 0, len(files))}
	for _, file := range files {
		set.File = append(set.File, protodesc.ToFileDescriptorProto(file))
	}
	return marshalDescriptors(set)
}

func marshalDescriptors(set *descriptorpb.FileDescriptorSet) ([]byte, error) {
	// A linked FileDescriptorSet contains no required proto2 fields, so the
	// maintained protobuf implementation cannot reject this concrete message.
	canonical, _ := (proto.MarshalOptions{Deterministic: true}).Marshal(set)
	return canonical, nil
}
