package schemaregistry

import (
	"context"
	"errors"
)

// privateDiagnostic keeps default formatting independent of collaborator text.
// Causes remain available only through deliberate error inspection.
func privateDiagnostic(message string, causes ...error) error {
	if len(causes) == 1 {
		return diagnosticError{message: message, cause: causes[0]}
	}
	return multiDiagnosticError{message: message, causes: causes}
}

// Only formerly bare passthrough boundaries preserve bare standard contexts.
func privatePassthroughDiagnostic(message string, cause error) error {
	if cause == context.Canceled || cause == context.DeadlineExceeded {
		return cause
	}
	private := passthroughDiagnostic{message: message, original: cause}
	if _, ok := cause.(interface{ Unwrap() []error }); ok {
		return multiPassthroughDiagnostic{private}
	}
	if _, ok := cause.(interface{ Unwrap() error }); ok {
		return singlePassthroughDiagnostic{private}
	}
	return private
}

// A formatting view keeps a formerly passed-through error's immediate topology.
// Original collaborator identity is retained only for explicit Is/As inspection.
type passthroughDiagnostic struct {
	message  string
	original error
}

func (err passthroughDiagnostic) Error() string        { return err.message }
func (err passthroughDiagnostic) Is(target error) bool { return errors.Is(err.original, target) }
func (err passthroughDiagnostic) As(target any) bool   { return errors.As(err.original, target) }

type singlePassthroughDiagnostic struct{ passthroughDiagnostic }

func (err singlePassthroughDiagnostic) Unwrap() error { return errors.Unwrap(err.original) }

type multiPassthroughDiagnostic struct{ passthroughDiagnostic }

func (err multiPassthroughDiagnostic) Unwrap() []error {
	return err.original.(interface{ Unwrap() []error }).Unwrap()
}

type multiDiagnosticError struct {
	message string
	causes  []error
}

func (err multiDiagnosticError) Error() string   { return err.message }
func (err multiDiagnosticError) Unwrap() []error { return err.causes }

type diagnosticError struct {
	message string
	cause   error
}

func (err diagnosticError) Error() string { return err.message }
func (err diagnosticError) Unwrap() error { return err.cause }

var (
	// ErrUnauthorized marks authentication or authorization rejection.
	ErrUnauthorized = errors.New("schema registry: unauthorized")
	// ErrIncompatible marks provider-enforced schema incompatibility.
	ErrIncompatible = errors.New("schema registry: incompatible")
	// ErrRejected marks a definitive provider rejection other than
	// incompatibility or authorization.
	ErrRejected = errors.New("schema registry: rejected")
	// ErrUnknownOutcome marks an operation whose effect cannot be determined.
	ErrUnknownOutcome = errors.New("schema registry: unknown outcome")
)
