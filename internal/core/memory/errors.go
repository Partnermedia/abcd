// Package memory is abcd's transport-agnostic curated-knowledge substrate at
// .abcd/memory/ (itd-36 / adr-13). Every verb — ingest, ask, lint, and the
// bare status render — is a function taking a structured request and returning
// a structured result plus error; nothing here writes stdout or knows about a
// CLI, MCP, or prompt surface. Surfaces under internal/surface/* marshal these
// results for their transport.
//
// WritePages is the ONE mutating entry to the store; it honours the ADR-13
// single-writer, atomic-rename crash model (advisory flock + six-step durable
// write + idempotent sibling reconciliation, no journal). Read-only paths
// (Ask, Bare, and the read half of Ingest) never heal drift — they report it.
package memory

import (
	"errors"
	"fmt"
	"os"
)

// bareErr strips a filesystem error down to its reason, dropping any absolute
// path a *PathError/*LinkError would otherwise re-embed in machine output
// (iss-81). The caller renders the path itself, repo-relative.
func bareErr(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	var le *os.LinkError
	if errors.As(err, &le) {
		return le.Err
	}
	return err
}

// IngestError is a pre-dispatch ingest failure — raised BEFORE any
// memory-store write (bad source path, fetch failure, binary source, zero
// distilled pages, repair collision). The surface reports it fail-closed and
// exits 1.
type IngestError struct{ Msg string }

func (e *IngestError) Error() string { return e.Msg }

func newIngestError(format string, a ...any) *IngestError {
	return &IngestError{Msg: fmt.Sprintf(format, a...)}
}

// AskError is a pre-write ask failure (unusable file-back payload, no matches
// to file back against, cited pages without complete provenance). Raised
// BEFORE any write; the surface reports it and exits 1.
type AskError struct{ Msg string }

func (e *AskError) Error() string { return e.Msg }

func newAskError(format string, a ...any) *AskError {
	return &AskError{Msg: fmt.Sprintf(format, a...)}
}

// WriterContractError is a writer-side contract violation (invalid write
// request, unreadable user-visible file the writer refuses to overwrite). No
// artifact is produced.
type WriterContractError struct{ Msg string }

func (e *WriterContractError) Error() string { return e.Msg }

func newWriterContractError(format string, a ...any) *WriterContractError {
	return &WriterContractError{Msg: fmt.Sprintf(format, a...)}
}

// StoreLockHeldError signals a live process holds .abcd/memory/.lock — the
// writer fails closed (LOCK_NB; the lock file is never deleted, flock releases
// on process exit).
type StoreLockHeldError struct{ Path string }

func (e *StoreLockHeldError) Error() string {
	return fmt.Sprintf("memory store lock is held by a live process: %s", e.Path)
}

// UnsafeStorePathError signals a memory-store path (.abcd, .abcd/memory, or the
// lock leaf) is a symlink or non-regular filesystem object.
type UnsafeStorePathError struct{ Msg string }

func (e *UnsafeStorePathError) Error() string { return e.Msg }

// MemorySchemaError signals a distilled page / source block that does not
// conform to the memory schema. No artifact is produced.
type MemorySchemaError struct{ Msg string }

func (e *MemorySchemaError) Error() string { return e.Msg }

func newSchemaError(format string, a ...any) *MemorySchemaError {
	return &MemorySchemaError{Msg: fmt.Sprintf(format, a...)}
}

// RegistryFormatError signals .sources_index.json exists but is not parseable
// as a JSON object — durable metadata must fail loudly, never be silently
// replaced with an empty index.
type RegistryFormatError struct{ Msg string }

func (e *RegistryFormatError) Error() string { return e.Msg }
