package model

import (
	"github.com/jtejido/golucene/core/util"
)

const (
	DOCS_POSITIONS_ENUM_FLAG_OFF_SETS = 1
	DOCS_POSITIONS_ENUM_FLAG_PAYLOADS = 2
)

type DocsAndPositionsEnum interface {
	DocsEnum
	/** Returns the next position.  You should only call this
	 *  up to {@link DocsEnum#freq()} times else
	 *  the behavior is not defined.  If positions were not
	 *  indexed this will return -1; this only happens if
	 *  offsets were indexed and you passed needsOffset=true
	 *  when pulling the enum.  */
	NextPosition() (int, error)

	/** Returns start offset for the current position, or -1
	 *  if offsets were not indexed. */
	StartOffset() (int, error)

	/** Returns end offset for the current position, or -1 if
	 *  offsets were not indexed. */
	EndOffset() (int, error)

	/** Returns the payload at this position, or null if no
	 *  payload was indexed. You should not modify anything
	 *  (neither members of the returned BytesRef nor bytes
	 *  in the byte[]). */
	Payload() (*util.BytesRef, error)
}

type EverythingEnum = DocsAndPositionsEnum
