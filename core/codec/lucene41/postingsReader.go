package lucene41

import (
	"fmt"
	"github.com/jtejido/golucene/core/codec"
	. "github.com/jtejido/golucene/core/codec/spi"
	. "github.com/jtejido/golucene/core/index/model"
	. "github.com/jtejido/golucene/core/search/model"
	"github.com/jtejido/golucene/core/store"
	"github.com/jtejido/golucene/core/util"
)

// Lucene41PostingsReader.java

/*
Concrete class that reads docId (maybe frq,pos,offset,payload) list
with postings format.
*/
type Lucene41PostingsReader struct {
	docIn   store.IndexInput
	posIn   store.IndexInput
	payIn   store.IndexInput
	forUtil *ForUtil
	version int
}

func NewLucene41PostingsReader(dir store.Directory,
	fis FieldInfos, si *SegmentInfo,
	ctx store.IOContext, segmentSuffix string) (r PostingsReaderBase, err error) {

	// fmt.Println("Initializing Lucene41PostingsReader...")
	success := false
	var docIn, posIn, payIn store.IndexInput = nil, nil, nil
	defer func() {
		if !success {
			fmt.Println("Failed to initialize Lucene41PostingsReader.")
			util.CloseWhileSuppressingError(docIn, posIn, payIn)
		}
	}()

	docIn, err = dir.OpenInput(util.SegmentFileName(si.Name, segmentSuffix, LUCENE41_DOC_EXTENSION), ctx)
	if err != nil {
		return nil, err
	}
	var version int32
	version, err = codec.CheckHeader(docIn, LUCENE41_DOC_CODEC, LUCENE41_VERSION_START, LUCENE41_VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	forUtil, err := NewForUtilFrom(docIn)
	if err != nil {
		return nil, err
	}

	if version >= LUCENE41_VERSION_CHECKSUM {
		// NOTE: data file is too costly to verify checksum against all the
		// bytes on open, but for now we at least verify proper structure
		// of the checksum footer: which looks for FOOTER_MAGIC +
		// algorithmID. This is cheap and can detect some forms of
		// corruption such as file trucation.
		if _, err = codec.RetrieveChecksum(docIn); err != nil {
			return nil, err
		}
	}

	if fis.HasProx {
		posIn, err = dir.OpenInput(util.SegmentFileName(si.Name, segmentSuffix, LUCENE41_POS_EXTENSION), ctx)
		if err != nil {
			return nil, err
		}
		_, err = codec.CheckHeader(posIn, LUCENE41_POS_CODEC, version, version)
		if err != nil {
			return nil, err
		}

		if version >= LUCENE41_VERSION_CHECKSUM {
			// NOTE: data file is too costly to verify checksum against all the
			// bytes on open, but for now we at least verify proper structure
			// of the checksum footer: which looks for FOOTER_MAGIC +
			// algorithmID. This is cheap and can detect some forms of
			// corruption such as file trucation.
			if _, err = codec.RetrieveChecksum(posIn); err != nil {
				return nil, err
			}
		}

		if fis.HasPayloads || fis.HasOffsets {
			payIn, err = dir.OpenInput(util.SegmentFileName(si.Name, segmentSuffix, LUCENE41_PAY_EXTENSION), ctx)
			if err != nil {
				return nil, err
			}
			_, err = codec.CheckHeader(payIn, LUCENE41_PAY_CODEC, version, version)
			if err != nil {
				return nil, err
			}

			if version >= LUCENE41_VERSION_CHECKSUM {
				// NOTE: data file is too costly to verify checksum against all the
				// bytes on open, but for now we at least verify proper structure
				// of the checksum footer: which looks for FOOTER_MAGIC +
				// algorithmID. This is cheap and can detect some forms of
				// corruption such as file trucation.
				if _, err = codec.RetrieveChecksum(payIn); err != nil {
					return nil, err
				}

			}
		}
	}

	success = true
	return &Lucene41PostingsReader{docIn, posIn, payIn, forUtil, int(version)}, nil
}

func (r *Lucene41PostingsReader) Init(termsIn store.IndexInput) error {
	// fmt.Println("Initializing from:", termsIn)
	// Make sure we are talking to the matching postings writer
	_, err := codec.CheckHeader(termsIn, LUCENE41_TERMS_CODEC, LUCENE41_VERSION_START, LUCENE41_VERSION_CURRENT)
	if err != nil {
		return err
	}
	indexBlockSize, err := termsIn.ReadVInt()
	if err != nil {
		return err
	}
	// fmt.Println("Index block size:", indexBlockSize)
	if indexBlockSize != LUCENE41_BLOCK_SIZE {
		panic(fmt.Sprintf("index-time BLOCK_SIZE (%v) != read-time BLOCK_SIZE (%v)", indexBlockSize, LUCENE41_BLOCK_SIZE))
	}
	return nil
}

/**
 * Read values that have been written using variable-length encoding instead of bit-packing.
 */
func readVIntBlock(docIn store.IndexInput, docBuffer []int32,
	freqBuffer []int32, num int, indexHasFreq bool) (err error) {
	if indexHasFreq {
		for i := 0; i < num; i++ {
			code, err := docIn.ReadVInt()
			if err != nil {
				return err
			}
			docBuffer[i] = int32(uint(code) >> 1)
			if (code & 1) != 0 {
				freqBuffer[i] = 1
			} else {
				freqBuffer[i], err = docIn.ReadVInt()
				if err != nil {
					return err
				}
			}
		}
	} else {
		for i := 0; i < num; i++ {
			docBuffer[i], err = docIn.ReadVInt()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func asInt(n int32, err error) (int, error) {
	return int(n), err
}

func (r *Lucene41PostingsReader) NewTermState() *BlockTermState {
	return newIntBlockTermState().BlockTermState
}

func (r *Lucene41PostingsReader) Close() error {
	return util.Close(r.docIn, r.posIn, r.payIn)
}

func (r *Lucene41PostingsReader) DecodeTerm(longs []int64,
	in util.DataInput, fieldInfo *FieldInfo,
	_termState *BlockTermState, absolute bool) (err error) {

	termState := _termState.Self.(*intBlockTermState)
	fieldHasPositions := fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
	fieldHasOffsets := fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	fieldHasPayloads := fieldInfo.HasPayloads()

	if absolute {
		termState.docStartFP = 0
		termState.posStartFP = 0
		termState.payStartFP = 0
	}
	if r.version < LUCENE41_VERSION_META_ARRAY { // backward compatibility
		return r._decodeTerm(in, fieldInfo, termState)
	}
	termState.docStartFP += longs[0]
	if fieldHasPositions {
		termState.posStartFP += longs[1]
		if fieldHasOffsets || fieldHasPayloads {
			termState.payStartFP += longs[2]
		}
	}
	if termState.DocFreq == 1 {
		if termState.singletonDocID, err = asInt(in.ReadVInt()); err != nil {
			return
		}
	} else {
		termState.singletonDocID = -1
	}
	if fieldHasPositions {
		if termState.TotalTermFreq > LUCENE41_BLOCK_SIZE {
			if termState.lastPosBlockOffset, err = in.ReadVLong(); err != nil {
				return err
			}
		} else {
			termState.lastPosBlockOffset = -1
		}
	}
	if termState.DocFreq > LUCENE41_BLOCK_SIZE {
		if termState.skipOffset, err = in.ReadVLong(); err != nil {
			return
		}
	} else {
		termState.skipOffset = -1
	}
	return nil
}

func (r *Lucene41PostingsReader) _decodeTerm(in util.DataInput,
	fieldInfo *FieldInfo, termState *intBlockTermState) (err error) {

	fieldHasPositions := fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
	fieldHasOffsets := fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	fieldHasPaylods := fieldInfo.HasPayloads()
	if termState.DocFreq == 1 {
		if termState.singletonDocID, err = asInt(in.ReadVInt()); err != nil {
			return
		}
	} else {
		termState.singletonDocID = -1
		var n int64
		if n, err = in.ReadVLong(); err != nil {
			return
		}
		termState.docStartFP += n
	}
	if fieldHasPositions {
		var n int64
		if n, err = in.ReadVLong(); err != nil {
			return
		}
		termState.posStartFP += n
		if termState.TotalTermFreq > LUCENE41_BLOCK_SIZE {
			if n, err = in.ReadVLong(); err != nil {
				return
			}
			termState.lastPosBlockOffset += n
		} else {
			termState.lastPosBlockOffset = -1
		}
		if (fieldHasPaylods || fieldHasOffsets) && termState.TotalTermFreq >= LUCENE41_BLOCK_SIZE {
			if n, err = in.ReadVLong(); err != nil {
				return
			}
			termState.payStartFP += n
		}
	}
	if termState.DocFreq > LUCENE41_BLOCK_SIZE {
		if termState.skipOffset, err = in.ReadVLong(); err != nil {
			return
		}
	} else {
		termState.skipOffset = -1
	}
	return nil
}

func (r *Lucene41PostingsReader) Docs(fieldInfo *FieldInfo,
	termState *BlockTermState, liveDocs util.Bits,
	reuse DocsEnum, flags int) (de DocsEnum, err error) {

	var docsEnum *blockDocsEnum
	if v, ok := reuse.(*blockDocsEnum); ok {
		docsEnum = v
		if !docsEnum.canReuse(r.docIn, fieldInfo) {
			docsEnum = newBlockDocsEnum(r, fieldInfo)
		}
	} else {
		docsEnum = newBlockDocsEnum(r, fieldInfo)
	}
	return docsEnum.reset(liveDocs, termState.Self.(*intBlockTermState), flags)
}

func (r *Lucene41PostingsReader) DocsAndPositions(fieldInfo *FieldInfo, termState *BlockTermState, liveDocs util.Bits, reuse DocsAndPositionsEnum, flags int) (DocsAndPositionsEnum, error) {

	indexHasOffsets := fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	indexHasPayloads := fieldInfo.HasPayloads()

	if (!indexHasOffsets || (flags&DOCS_POSITIONS_ENUM_FLAG_OFF_SETS) == 0) &&
		(!indexHasPayloads || (flags&DOCS_POSITIONS_ENUM_FLAG_PAYLOADS) == 0) {

		var docsAndPositionsEnum *blockDocsAndPositionsEnum
		if v, ok := reuse.(*blockDocsAndPositionsEnum); ok {
			docsAndPositionsEnum = v
			if !docsAndPositionsEnum.canReuse(r.docIn, fieldInfo) {
				docsAndPositionsEnum = newBlockDocsAndPositionsEnum(r, fieldInfo)
			}
		} else {
			docsAndPositionsEnum = newBlockDocsAndPositionsEnum(r, fieldInfo)
		}

		return docsAndPositionsEnum.reset(liveDocs, termState.Self.(*intBlockTermState))
	}

	var e *everythingEnum
	if v, ok := reuse.(*everythingEnum); ok {
		e = v
		if !e.canReuse(r.docIn, fieldInfo) {
			e = newEverythingEnum(r, fieldInfo)
		}
	} else {
		e = newEverythingEnum(r, fieldInfo)
	}
	return e.reset(liveDocs, termState.Self.(*intBlockTermState), flags)

}

type blockDocsEnum struct {
	*Lucene41PostingsReader // embedded struct

	encoded []byte

	docDeltaBuffer []int32
	freqBuffer     []int32

	docBufferUpto int

	// skipper Lucene41SkipReader
	skipped bool

	startDocIn store.IndexInput

	docIn            store.IndexInput
	indexHasFreq     bool
	indexHasPos      bool
	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int
	totalTermFreq int64
	docUpto       int
	doc           int
	accum         int
	freq          int

	// Where this term's postings start in the .doc file:
	docTermStartFP int64

	// Where this term's skip data starts (after
	// docTermStartFP) in the .doc file (or -1 if there is
	// no skip data for this term):
	skipOffset int64

	// docID for next skip point, we won't use skipper if
	// target docID is not larger than this
	nextSkipDoc int

	liveDocs util.Bits

	needsFreq      bool
	singletonDocID int
}

func newBlockDocsEnum(owner *Lucene41PostingsReader,
	fieldInfo *FieldInfo) *blockDocsEnum {

	return &blockDocsEnum{
		Lucene41PostingsReader: owner,
		docDeltaBuffer:         make([]int32, MAX_DATA_SIZE),
		freqBuffer:             make([]int32, MAX_DATA_SIZE),
		startDocIn:             owner.docIn,
		docIn:                  nil,
		indexHasFreq:           fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS,
		indexHasPos:            fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS,
		indexHasOffsets:        fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS,
		indexHasPayloads:       fieldInfo.HasPayloads(),
		encoded:                make([]byte, MAX_ENCODED_SIZE),
	}
}

func (de *blockDocsEnum) canReuse(docIn store.IndexInput, fieldInfo *FieldInfo) bool {
	return docIn == de.startDocIn &&
		de.indexHasFreq == (fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS) &&
		de.indexHasPos == (fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS) &&
		de.indexHasPayloads == fieldInfo.HasPayloads()
}

func (de *blockDocsEnum) reset(liveDocs util.Bits, termState *intBlockTermState, flags int) (ret DocsEnum, err error) {
	de.liveDocs = liveDocs
	// fmt.Println("  FPR.reset: termState=", termState)
	de.docFreq = termState.DocFreq
	if de.indexHasFreq {
		de.totalTermFreq = termState.TotalTermFreq
	} else {
		de.totalTermFreq = int64(de.docFreq)
	}
	de.docTermStartFP = termState.docStartFP // <---- docTermStartFP should be 178 instead of 0
	de.skipOffset = termState.skipOffset
	de.singletonDocID = termState.singletonDocID
	if de.docFreq > 1 {
		if de.docIn == nil {
			// lazy init
			de.docIn = de.startDocIn.Clone()
		}
		err = de.docIn.Seek(de.docTermStartFP)
		if err != nil {
			return nil, err
		}
	}

	de.doc = -1
	de.needsFreq = (flags & DOCS_ENUM_FLAG_FREQS) != 0
	if !de.indexHasFreq {
		for i, _ := range de.freqBuffer {
			de.freqBuffer[i] = 1
		}
	}
	de.accum = 0
	de.docUpto = 0
	de.nextSkipDoc = LUCENE41_BLOCK_SIZE - 1 // we won't skip if target is found in first block
	de.docBufferUpto = LUCENE41_BLOCK_SIZE
	de.skipped = false
	return de, nil
}

func (de *blockDocsEnum) Freq() (n int, err error) {
	return de.freq, nil
}

func (de *blockDocsEnum) DocId() int {
	return de.doc
}

func (de *blockDocsEnum) refillDocs() (err error) {
	left := de.docFreq - de.docUpto
	assert(left > 0)

	if left >= LUCENE41_BLOCK_SIZE {
		fmt.Println("    fill doc block from fp=", de.docIn.FilePointer())
		panic("not implemented yet")
	} else if de.docFreq == 1 {
		de.docDeltaBuffer[0] = int32(de.singletonDocID)
		de.freqBuffer[0] = int32(de.totalTermFreq)
	} else {
		// Read vInts:
		// fmt.Println("    fill last vInt block from fp=", de.docIn.FilePointer())
		err = readVIntBlock(de.docIn, de.docDeltaBuffer, de.freqBuffer, left, de.indexHasFreq)
	}
	de.docBufferUpto = 0
	return
}

func (de *blockDocsEnum) NextDoc() (n int, err error) {
	// fmt.Println("FPR.nextDoc")
	for {
		// fmt.Printf("  docUpto=%v (of df=%v) docBufferUpto=%v\n", de.docUpto, de.docFreq, de.docBufferUpto)

		if de.docUpto == de.docFreq {
			// fmt.Println("  return doc=END")
			de.doc = NO_MORE_DOCS
			return de.doc, nil
		}

		if de.docBufferUpto == LUCENE41_BLOCK_SIZE {
			err = de.refillDocs()
			if err != nil {
				return 0, err
			}
		}

		// fmt.Printf("    accum=%v docDeltaBuffer[%v]=%v\n", de.accum, de.docBufferUpto, de.docDeltaBuffer[de.docBufferUpto])
		de.accum += int(de.docDeltaBuffer[de.docBufferUpto])
		de.docUpto++

		if de.liveDocs == nil || de.liveDocs.At(de.accum) {
			de.doc = de.accum
			de.freq = int(de.freqBuffer[de.docBufferUpto])
			de.docBufferUpto++
			// fmt.Printf("  return doc=%v freq=%v\n", de.doc, de.freq)
			return de.doc, nil
		}
		// fmt.Printf("  doc=%v is deleted; try next doc\n", de.accum)
		de.docBufferUpto++
	}
}

func (de *blockDocsEnum) Advance(target int) (int, error) {
	// TODO: make frq block load lazy/skippable
	fmt.Printf("  FPR.advance target=%v\n", target)

	// current skip docID < docIDs generated from current buffer <= next
	// skip docID, we don't need to skip if target is buffered already
	if de.docFreq > LUCENE41_BLOCK_SIZE && target > de.nextSkipDoc {
		fmt.Println("load skipper")

		panic("not implemented yet")
	}
	if de.docUpto == de.docFreq {
		de.doc = NO_MORE_DOCS
		return de.doc, nil
	}
	if de.docBufferUpto == LUCENE41_BLOCK_SIZE {
		err := de.refillDocs()
		if err != nil {
			return 0, nil
		}
	}

	// Now scan.. this is an inlined/pared down version of nextDoc():
	for {
		fmt.Printf("  scan doc=%v docBufferUpto=%v\n", de.accum, de.docBufferUpto)
		de.accum += int(de.docDeltaBuffer[de.docBufferUpto])
		de.docUpto++

		if de.accum >= target {
			break
		}
		de.docBufferUpto++
		if de.docUpto == de.docFreq {
			de.doc = NO_MORE_DOCS
			return de.doc, nil
		}
	}

	if de.liveDocs == nil || de.liveDocs.At(de.accum) {
		fmt.Printf("  return doc=%v\n", de.accum)
		de.freq = int(de.freqBuffer[de.docBufferUpto])
		de.docBufferUpto++
		de.doc = de.accum
		return de.doc, nil
	} else {
		fmt.Println("  now do nextDoc()")
		de.docBufferUpto++
		return de.NextDoc()
	}
}

func (de *blockDocsEnum) Cost() int64 {
	return int64(de.docFreq)
}

type blockDocsAndPositionsEnum struct {
	*Lucene41PostingsReader // embedded struct

	encoded []byte

	docDeltaBuffer []int32
	freqBuffer     []int32
	posDeltaBuffer []int32

	docBufferUpto, posBufferUpto int

	// skipper Lucene41SkipReader
	skipped bool

	startDocIn store.IndexInput

	docIn, posIn     store.IndexInput
	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int
	totalTermFreq int64
	docUpto       int
	doc           int
	accum         int
	freq          int
	position      int

	// how many positions "behind" we are; nextPosition must
	// skip these to "catch up":
	posPendingCount int

	// Lazy pos seek: if != -1 then we must seek to this FP
	// before reading positions:
	posPendingFP int64
	// Where this term's postings start in the .doc file:
	docTermStartFP int64

	// Where this term's postings start in the .pos file:
	posTermStartFP int64

	// Where this term's payloads/offsets start in the .pay
	// file:
	payTermStartFP int64

	// File pointer where the last (vInt encoded) pos delta
	// block is.  We need this to know whether to bulk
	// decode vs vInt decode the block:
	lastPosBlockFP int64

	// Where this term's skip data starts (after
	// docTermStartFP) in the .doc file (or -1 if there is
	// no skip data for this term):
	skipOffset int64

	// docID for next skip point, we won't use skipper if
	// target docID is not larger than this
	nextSkipDoc int

	liveDocs util.Bits

	singletonDocID int
}

func newBlockDocsAndPositionsEnum(owner *Lucene41PostingsReader, fieldInfo *FieldInfo) *blockDocsAndPositionsEnum {
	return &blockDocsAndPositionsEnum{
		Lucene41PostingsReader: owner,
		docDeltaBuffer:         make([]int32, MAX_DATA_SIZE),
		freqBuffer:             make([]int32, MAX_DATA_SIZE),
		posDeltaBuffer:         make([]int32, MAX_DATA_SIZE),
		startDocIn:             owner.docIn,
		docIn:                  nil,
		posIn:                  owner.posIn.Clone(),
		indexHasOffsets:        fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS,
		indexHasPayloads:       fieldInfo.HasPayloads(),
		encoded:                make([]byte, MAX_ENCODED_SIZE),
	}
}

func (de *blockDocsAndPositionsEnum) canReuse(docIn store.IndexInput, fieldInfo *FieldInfo) bool {
	return docIn == de.startDocIn &&
		de.indexHasOffsets == (fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS) &&
		de.indexHasPayloads == fieldInfo.HasPayloads()
}

func (de *blockDocsAndPositionsEnum) reset(liveDocs util.Bits, termState *intBlockTermState) (ret DocsAndPositionsEnum, err error) {
	de.liveDocs = liveDocs
	// fmt.Println("  FPR.reset: termState=", termState)
	de.docFreq = termState.DocFreq
	de.docTermStartFP = termState.docStartFP
	de.posTermStartFP = termState.posStartFP
	de.payTermStartFP = termState.payStartFP
	de.skipOffset = termState.skipOffset
	de.totalTermFreq = termState.TotalTermFreq
	de.singletonDocID = termState.singletonDocID

	if de.docFreq > 1 {
		if de.docIn == nil {
			// lazy init
			de.docIn = de.startDocIn.Clone()
		}
		err = de.docIn.Seek(de.docTermStartFP)
		if err != nil {
			return nil, err
		}
	}
	de.posPendingFP = de.posTermStartFP
	de.posPendingCount = 0

	if termState.TotalTermFreq < LUCENE41_BLOCK_SIZE {
		de.lastPosBlockFP = de.posTermStartFP
	} else if termState.TotalTermFreq == LUCENE41_BLOCK_SIZE {
		de.lastPosBlockFP = -1
	} else {
		de.lastPosBlockFP = de.posTermStartFP + termState.lastPosBlockOffset
	}

	de.doc = -1
	de.accum = 0
	de.docUpto = 0
	if de.docFreq > LUCENE41_BLOCK_SIZE {
		de.nextSkipDoc = LUCENE41_BLOCK_SIZE - 1 // we won't skip if target is found in first block
	} else {
		de.nextSkipDoc = NO_MORE_DOCS // not enough docs for skipping
	}
	de.docBufferUpto = LUCENE41_BLOCK_SIZE
	de.skipped = false

	return de, nil
}

func (de *blockDocsAndPositionsEnum) Freq() (n int, err error) {
	return de.freq, nil
}

func (de *blockDocsAndPositionsEnum) DocId() int {
	return de.doc
}

func (de *blockDocsAndPositionsEnum) refillDocs() (err error) {
	left := de.docFreq - de.docUpto
	assert(left > 0)

	if left >= LUCENE41_BLOCK_SIZE {
		// if (DEBUG) {
		//   System.out.println("    fill doc block from fp=" + docIn.getFilePointer());
		// }
		panic("not implemented yet")
		// err = de.Lucene41PostingsReader.forUtil.readBlock(de.docIn, de.encoded, de.docDeltaBuffer)
		// if err != nil {
		// 	return
		// }
		// if (DEBUG) {
		//   System.out.println("    fill freq block from fp=" + docIn.getFilePointer());
		// }
		// err = de.Lucene41PostingsReader.forUtil.readBlock(de.docIn, de.encoded, de.freqBuffer)
		// if err != nil {
		// 	return
		// }
	} else if de.docFreq == 1 {
		de.docDeltaBuffer[0] = int32(de.singletonDocID)
		de.freqBuffer[0] = int32(de.totalTermFreq)
	} else {
		// Read vInts:
		// if (DEBUG) {
		//   System.out.println("    fill last vInt doc block from fp=" + docIn.getFilePointer());
		// }
		err = readVIntBlock(de.docIn, de.docDeltaBuffer, de.freqBuffer, left, true)
		if err != nil {
			return
		}
	}
	de.docBufferUpto = 0
	return
}

func (de *blockDocsAndPositionsEnum) refillPositions() (err error) {
	// if (DEBUG) {
	//   System.out.println("      refillPositions");
	// }
	if de.posIn.FilePointer() == de.lastPosBlockFP {
		// if (DEBUG) {
		//   System.out.println("        vInt pos block @ fp=" + posIn.getFilePointer() + " hasPayloads=" + indexHasPayloads + " hasOffsets=" + indexHasOffsets);
		// }
		count := int(de.totalTermFreq % LUCENE41_BLOCK_SIZE)
		var payloadLength int32
		for i := 0; i < count; i++ {
			code, err := de.posIn.ReadVInt()
			if err != nil {
				return err
			}
			if de.indexHasPayloads {
				if (code & 1) != 0 {
					payloadLength, err = de.posIn.ReadVInt()
					if err != nil {
						return err
					}
				}
				de.posDeltaBuffer[i] = code >> 1
				if payloadLength != 0 {
					de.posIn.Seek(de.posIn.FilePointer() + int64(payloadLength))
				}
			} else {
				de.posDeltaBuffer[i] = code
			}
			if de.indexHasOffsets {
				v, err := de.posIn.ReadVInt()
				if err != nil {
					return err
				}
				if (v & 1) != 0 {
					// offset length changed
					_, err = de.posIn.ReadVInt()
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		// if (DEBUG) {
		//   System.out.println("        bulk pos block @ fp=" + posIn.getFilePointer());
		// }
		err = de.forUtil.readBlock(de.posIn, de.encoded, de.posDeltaBuffer)
	}

	return
}

func (de *blockDocsAndPositionsEnum) skipPositions() (err error) {
	// Skip positions now:
	toSkip := de.posPendingCount - de.freq
	// if (DEBUG) {
	//   System.out.println("      FPR.skipPositions: toSkip=" + toSkip);
	// }

	leftInBlock := LUCENE41_BLOCK_SIZE - de.posBufferUpto
	if toSkip < leftInBlock {
		de.posBufferUpto += toSkip
		// if (DEBUG) {
		//   System.out.println("        skip w/in block to posBufferUpto=" + posBufferUpto);
		// }
	} else {
		toSkip -= leftInBlock
		for toSkip >= LUCENE41_BLOCK_SIZE {
			// if (DEBUG) {
			//   System.out.println("        skip whole block @ fp=" + posIn.getFilePointer());
			// }
			// assert(de.posIn.FilePointer() != de.lastPosBlockFP)
			// de.forUtil.skipBlock(de.posIn)
			// toSkip -= LUCENE41_BLOCK_SIZE
			panic("niy")
		}
		if err = de.refillPositions(); err != nil {
			return
		}
		de.posBufferUpto = toSkip
		// if (DEBUG) {
		//   System.out.println("        skip w/in block to posBufferUpto=" + posBufferUpto);
		// }
	}

	de.position = 0
	return nil
}

func (de *blockDocsAndPositionsEnum) NextDoc() (n int, err error) {
	// if (DEBUG) {
	//   System.out.println("  FPR.nextDoc");
	// }
	for {
		// if (DEBUG) {
		//   System.out.println("    docUpto=" + docUpto + " (of df=" + docFreq + ") docBufferUpto=" + docBufferUpto);
		// }
		if de.docUpto == de.docFreq {
			de.doc = NO_MORE_DOCS
			return de.doc, nil
		}
		if de.docBufferUpto == LUCENE41_BLOCK_SIZE {
			if err = de.refillDocs(); err != nil {
				return 0, err
			}
		}
		// if (DEBUG) {
		//   System.out.println("    accum=" + accum + " docDeltaBuffer[" + docBufferUpto + "]=" + docDeltaBuffer[docBufferUpto]);
		// }
		de.accum += int(de.docDeltaBuffer[de.docBufferUpto])
		de.freq = int(de.freqBuffer[de.docBufferUpto])
		de.posPendingCount += de.freq
		de.docBufferUpto++
		de.docUpto++

		if de.liveDocs == nil || de.liveDocs.At(de.accum) {
			de.doc = de.accum
			de.position = 0
			// if (DEBUG) {
			//   System.out.println("    return doc=" + doc + " freq=" + freq + " posPendingCount=" + posPendingCount);
			// }
			return de.doc, nil
		}
		// if (DEBUG) {
		//   System.out.println("    doc=" + accum + " is deleted; try next doc");
		// }
	}
}

func (de *blockDocsAndPositionsEnum) Advance(target int) (int, error) {
	// TODO: make frq block load lazy/skippable
	// if (DEBUG) {
	//   System.out.println("  FPR.advance target=" + target);
	// }

	if target > de.nextSkipDoc {
		panic("not implemented yet")
	}
	if de.docUpto == de.docFreq {
		de.doc = NO_MORE_DOCS
		return de.doc, nil
	}
	if de.docBufferUpto == LUCENE41_BLOCK_SIZE {
		de.refillDocs()
	}

	// Now scan... this is an inlined/pared down version
	// of nextDoc():
	for {
		// if (DEBUG) {
		//   System.out.println("  scan doc=" + accum + " docBufferUpto=" + docBufferUpto);
		// }
		de.accum += int(de.docDeltaBuffer[de.docBufferUpto])
		de.freq = int(de.freqBuffer[de.docBufferUpto])
		de.posPendingCount += de.freq
		de.docBufferUpto++
		de.docUpto++

		if de.accum >= target {
			break
		}
		if de.docUpto == de.docFreq {
			de.doc = NO_MORE_DOCS
			return de.doc, nil
		}
	}

	if de.liveDocs == nil || de.liveDocs.At(de.accum) {
		// if (DEBUG) {
		//   System.out.println("  return doc=" + accum);
		// }
		de.position = 0
		de.doc = de.accum
		return de.doc, nil
	} else {
		// if (DEBUG) {
		//   System.out.println("  now do nextDoc()");
		// }
		return de.NextDoc()
	}
}

func (de *blockDocsAndPositionsEnum) NextPosition() (pos int, err error) {
	// if (DEBUG) {
	//   System.out.println("    FPR.nextPosition posPendingCount=" + posPendingCount + " posBufferUpto=" + posBufferUpto);
	// }
	if de.posPendingFP != -1 {
		// if (DEBUG) {
		//   System.out.println("      seek to pendingFP=" + posPendingFP);
		// }
		de.posIn.Seek(de.posPendingFP)
		de.posPendingFP = -1

		// Force buffer refill:
		de.posBufferUpto = LUCENE41_BLOCK_SIZE
	}

	if de.posPendingCount > de.freq {
		de.skipPositions()
		de.posPendingCount = de.freq
	}

	if de.posBufferUpto == LUCENE41_BLOCK_SIZE {
		if err = de.refillPositions(); err != nil {
			return
		}
		de.posBufferUpto = 0
	}
	de.position += int(de.posDeltaBuffer[de.posBufferUpto])
	de.posBufferUpto++
	de.posPendingCount--
	// if (DEBUG) {
	//   System.out.println("      return pos=" + position);
	// }
	return de.position, nil
}

func (de *blockDocsAndPositionsEnum) Cost() int64 {
	return int64(de.docFreq)
}

func (de *blockDocsAndPositionsEnum) StartOffset() (int, error) {
	return -1, nil
}

func (de *blockDocsAndPositionsEnum) EndOffset() (int, error) {
	return -1, nil
}

func (de *blockDocsAndPositionsEnum) Payload() (*util.BytesRef, error) {
	return nil, nil
}

// Also handles payloads + offsets
type everythingEnum struct {
	*Lucene41PostingsReader // embedded struct

	encoded []byte

	docDeltaBuffer         []int32
	freqBuffer             []int32
	posDeltaBuffer         []int32
	payloadLengthBuffer    []int32
	offsetStartDeltaBuffer []int32
	offsetLengthBuffer     []int32
	payloadBytes           []byte

	payloadByteUpto, payloadLength, lastStartOffset, startOffset, endOffset, docBufferUpto, posBufferUpto int

	// skipper Lucene41SkipReader
	skipped bool

	startDocIn store.IndexInput

	docIn, posIn, payIn store.IndexInput

	payload *util.BytesRef

	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int
	totalTermFreq int64
	docUpto       int
	doc           int
	accum         int
	freq          int
	position      int

	// how many positions "behind" we are; nextPosition must
	// skip these to "catch up":
	posPendingCount int

	// Lazy pos seek: if != -1 then we must seek to this FP
	// before reading positions:
	posPendingFP int64

	// Lazy pay seek: if != -1 then we must seek to this FP
	// before reading payloads/offsets:
	payPendingFP int64

	// Where this term's postings start in the .doc file:
	docTermStartFP int64

	// Where this term's postings start in the .pos file:
	posTermStartFP int64

	// Where this term's payloads/offsets start in the .pay
	// file:
	payTermStartFP int64

	// File pointer where the last (vInt encoded) pos delta
	// block is.  We need this to know whether to bulk
	// decode vs vInt decode the block:
	lastPosBlockFP int64

	// Where this term's skip data starts (after
	// docTermStartFP) in the .doc file (or -1 if there is
	// no skip data for this term):
	skipOffset int64

	// docID for next skip point, we won't use skipper if
	// target docID is not larger than this
	nextSkipDoc int

	liveDocs util.Bits

	singletonDocID int

	needsOffsets, needsPayloads bool
}

func newEverythingEnum(owner *Lucene41PostingsReader, fieldInfo *FieldInfo) *everythingEnum {
	ans := &everythingEnum{
		Lucene41PostingsReader: owner,
		docDeltaBuffer:         make([]int32, MAX_DATA_SIZE),
		freqBuffer:             make([]int32, MAX_DATA_SIZE),
		posDeltaBuffer:         make([]int32, MAX_DATA_SIZE),
		startDocIn:             owner.docIn,
		docIn:                  nil,
		posIn:                  owner.posIn.Clone(),
		payIn:                  owner.payIn.Clone(),
		indexHasOffsets:        fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS,
		indexHasPayloads:       fieldInfo.HasPayloads(),
		encoded:                make([]byte, MAX_ENCODED_SIZE),
	}

	if ans.indexHasOffsets {
		ans.offsetStartDeltaBuffer = make([]int32, MAX_DATA_SIZE)
		ans.offsetLengthBuffer = make([]int32, MAX_DATA_SIZE)
	} else {
		ans.startOffset = -1
		ans.endOffset = -1
	}

	if ans.indexHasPayloads {
		ans.payloadLengthBuffer = make([]int32, MAX_DATA_SIZE)
		ans.payloadBytes = make([]byte, 128)
		ans.payload = util.NewEmptyBytesRef()
	}

	return ans
}

func (de *everythingEnum) canReuse(docIn store.IndexInput, fieldInfo *FieldInfo) bool {
	return de.docIn == de.startDocIn &&
		de.indexHasOffsets == (fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS) &&
		de.indexHasPayloads == fieldInfo.HasPayloads()
}

func (de *everythingEnum) reset(liveDocs util.Bits, termState *intBlockTermState, flags int) (ret EverythingEnum, err error) {
	de.liveDocs = liveDocs
	// if (DEBUG) {
	//   System.out.println("  FPR.reset: termState=" + termState);
	// }
	de.docFreq = termState.DocFreq
	de.docTermStartFP = termState.docStartFP
	de.posTermStartFP = termState.posStartFP
	de.payTermStartFP = termState.payStartFP
	de.skipOffset = termState.skipOffset
	de.totalTermFreq = termState.TotalTermFreq
	de.singletonDocID = termState.singletonDocID
	if de.docFreq > 1 {
		if de.docIn == nil {
			// lazy init
			de.docIn = de.startDocIn.Clone()
		}
		err = de.docIn.Seek(de.docTermStartFP)
		if err != nil {
			return nil, err
		}
	}
	de.posPendingFP = de.posTermStartFP
	de.payPendingFP = de.payTermStartFP
	de.posPendingCount = 0
	if termState.TotalTermFreq < LUCENE41_BLOCK_SIZE {
		de.lastPosBlockFP = de.posTermStartFP
	} else if termState.TotalTermFreq == LUCENE41_BLOCK_SIZE {
		de.lastPosBlockFP = -1
	} else {
		de.lastPosBlockFP = de.posTermStartFP + termState.lastPosBlockOffset
	}

	de.needsOffsets = (flags & DOCS_POSITIONS_ENUM_FLAG_OFF_SETS) != 0
	de.needsPayloads = (flags & DOCS_POSITIONS_ENUM_FLAG_OFF_SETS) != 0

	de.doc = -1
	de.accum = 0
	de.docUpto = 0
	if de.docFreq > LUCENE41_BLOCK_SIZE {
		de.nextSkipDoc = LUCENE41_BLOCK_SIZE - 1 // we won't skip if target is found in first block
	} else {
		de.nextSkipDoc = NO_MORE_DOCS // not enough docs for skipping
	}
	de.docBufferUpto = LUCENE41_BLOCK_SIZE
	de.skipped = false
	return de, nil
}

func (de *everythingEnum) Freq() (n int, err error) {
	return de.freq, nil
}

func (de *everythingEnum) DocId() int {
	return de.doc
}

func (de *everythingEnum) refillDocs() (err error) {
	left := de.docFreq - de.docUpto
	assert(left > 0)

	if left >= LUCENE41_BLOCK_SIZE {
		// if (DEBUG) {
		//   System.out.println("    fill doc block from fp=" + docIn.getFilePointer());
		// }
		if err = de.forUtil.readBlock(de.docIn, de.encoded, de.docDeltaBuffer); err != nil {
			return
		}
		// if (DEBUG) {
		//   System.out.println("    fill freq block from fp=" + docIn.getFilePointer());
		// }
		if err = de.forUtil.readBlock(de.docIn, de.encoded, de.freqBuffer); err != nil {
			return
		}
	} else if de.docFreq == 1 {
		de.docDeltaBuffer[0] = int32(de.singletonDocID)
		de.freqBuffer[0] = int32(de.totalTermFreq)
	} else {
		// if (DEBUG) {
		//   System.out.println("    fill last vInt doc block from fp=" + docIn.getFilePointer());
		// }
		if err = readVIntBlock(de.docIn, de.docDeltaBuffer, de.freqBuffer, left, true); err != nil {
			return
		}
	}
	de.docBufferUpto = 0
	return nil
}

func (de *everythingEnum) refillPositions() (err error) {
	// if (DEBUG) {
	//   System.out.println("      refillPositions");
	// }
	if de.posIn.FilePointer() == de.lastPosBlockFP {
		// if (DEBUG) {
		//   System.out.println("        vInt pos block @ fp=" + posIn.getFilePointer() + " hasPayloads=" + indexHasPayloads + " hasOffsets=" + indexHasOffsets);
		// }
		count := int(de.totalTermFreq % LUCENE41_BLOCK_SIZE)
		payloadLength := int32(0)
		offsetLength := int32(0)
		de.payloadByteUpto = 0
		for i := 0; i < count; i++ {
			code, err := de.posIn.ReadVInt()
			if err != nil {
				return err
			}
			if de.indexHasPayloads {
				if (code & 1) != 0 {
					if payloadLength, err = de.posIn.ReadVInt(); err != nil {
						return err
					}
				}
				// if (DEBUG) {
				//   System.out.println("        i=" + i + " payloadLen=" + payloadLength);
				// }
				de.payloadLengthBuffer[i] = payloadLength
				de.posDeltaBuffer[i] = code >> 1
				if de.payloadLength != 0 {
					if de.payloadByteUpto+int(payloadLength) > len(de.payloadBytes) {
						de.payloadBytes = util.GrowByteSlice(de.payloadBytes, de.payloadByteUpto+int(payloadLength))
					}
					//System.out.println("          read payload @ pos.fp=" + posIn.getFilePointer());
					if err = de.posIn.ReadBytes(de.payloadBytes[de.payloadByteUpto:payloadLength]); err != nil {
						return err
					}
					de.payloadByteUpto += int(payloadLength)
				}
			} else {
				de.posDeltaBuffer[i] = code
			}

			if de.indexHasOffsets {
				// if (DEBUG) {
				//   System.out.println("        i=" + i + " read offsets from posIn.fp=" + posIn.getFilePointer());
				// }
				deltaCode, err := de.posIn.ReadVInt()
				if err != nil {
					return err
				}
				if (deltaCode & 1) != 0 {
					if offsetLength, err = de.posIn.ReadVInt(); err != nil {
						return err
					}
				}
				de.offsetStartDeltaBuffer[i] = deltaCode >> 1
				de.offsetLengthBuffer[i] = offsetLength
				// if (DEBUG) {
				//   System.out.println("          startOffDelta=" + offsetStartDeltaBuffer[i] + " offsetLen=" + offsetLengthBuffer[i]);
				// }
			}
		}
		de.payloadByteUpto = 0
	} else {
		// if (DEBUG) {
		//   System.out.println("        bulk pos block @ fp=" + posIn.getFilePointer());
		// }
		if err = de.forUtil.readBlock(de.posIn, de.encoded, de.posDeltaBuffer); err != nil {
			return
		}

		if de.indexHasPayloads {
			// if (DEBUG) {
			//   System.out.println("        bulk payload block @ pay.fp=" + payIn.getFilePointer());
			// }
			if de.needsPayloads {
				if err = de.forUtil.readBlock(de.payIn, de.encoded, de.payloadLengthBuffer); err != nil {
					return
				}
				numBytes, err := de.payIn.ReadVInt()
				if err != nil {
					return err
				}
				// if (DEBUG) {
				//   System.out.println("        " + numBytes + " payload bytes @ pay.fp=" + payIn.getFilePointer());
				// }
				if int(numBytes) > len(de.payloadBytes) {
					de.payloadBytes = util.GrowByteSlice(de.payloadBytes, int(numBytes))
				}
				if err = de.payIn.ReadBytes(de.payloadBytes[:numBytes]); err != nil {
					return err
				}
			} else {
				// this works, because when writing a vint block we always force the first length to be written
				// de.forUtil.skipBlock(payIn); // skip over lengths
				// int numBytes = payIn.readVInt(); // read length of payloadBytes
				// payIn.seek(payIn.getFilePointer() + numBytes); // skip over payloadBytes
				panic("niy")
			}
			de.payloadByteUpto = 0
		}

		if de.indexHasOffsets {
			// if (DEBUG) {
			//   System.out.println("        bulk offset block @ pay.fp=" + payIn.getFilePointer());
			// }
			if de.needsOffsets {
				if err = de.forUtil.readBlock(de.payIn, de.encoded, de.offsetStartDeltaBuffer); err != nil {
					return
				}
				if err = de.forUtil.readBlock(de.payIn, de.encoded, de.offsetLengthBuffer); err != nil {
					return
				}
			} else {
				// this works, because when writing a vint block we always force the first length to be written
				// de.forUtil.skipBlock(payIn); // skip over starts
				// de.forUtil.skipBlock(payIn); // skip over lengths
				panic("niy")
			}
		}
	}

	return nil
}

func (de *everythingEnum) NextDoc() (n int, err error) {
	// if (DEBUG) {
	//   System.out.println("  FPR.nextDoc");
	// }
	for {
		// if (DEBUG) {
		//   System.out.println("    docUpto=" + docUpto + " (of df=" + docFreq + ") docBufferUpto=" + docBufferUpto);
		// }
		if de.docUpto == de.docFreq {
			de.doc = NO_MORE_DOCS
			return de.doc, nil
		}
		if de.docBufferUpto == LUCENE41_BLOCK_SIZE {
			de.refillDocs()
		}
		// if (DEBUG) {
		//   System.out.println("    accum=" + accum + " docDeltaBuffer[" + docBufferUpto + "]=" + docDeltaBuffer[docBufferUpto]);
		// }
		de.accum += int(de.docDeltaBuffer[de.docBufferUpto])
		de.freq = int(de.freqBuffer[de.docBufferUpto])
		de.posPendingCount += de.freq
		de.docBufferUpto++
		de.docUpto++

		if de.liveDocs == nil || de.liveDocs.At(de.accum) {
			de.doc = de.accum
			// if (DEBUG) {
			//   System.out.println("    return doc=" + doc + " freq=" + freq + " posPendingCount=" + posPendingCount);
			// }
			de.position = 0
			de.lastStartOffset = 0
			return de.doc, nil
		}

		// if (DEBUG) {
		//   System.out.println("    doc=" + accum + " is deleted; try next doc");
		// }
	}
}

func (de *everythingEnum) Advance(target int) (int, error) {
	// TODO: make frq block load lazy/skippable
	// if (DEBUG) {
	//   System.out.println("  FPR.advance target=" + target);
	// }

	if target > de.nextSkipDoc {

		panic("niy")
	}
	if de.docUpto == de.docFreq {
		de.doc = NO_MORE_DOCS
		return de.doc, nil
	}
	if de.docBufferUpto == LUCENE41_BLOCK_SIZE {
		de.refillDocs()
	}

	// Now scan:
	for {
		// if (DEBUG) {
		//   System.out.println("  scan doc=" + accum + " docBufferUpto=" + docBufferUpto);
		// }
		de.accum += int(de.docDeltaBuffer[de.docBufferUpto])
		de.freq = int(de.freqBuffer[de.docBufferUpto])
		de.posPendingCount += de.freq
		de.docBufferUpto++
		de.docUpto++

		if de.accum >= target {
			break
		}
		if de.docUpto == de.docFreq {
			de.doc = NO_MORE_DOCS
			return de.doc, nil
		}
	}

	if de.liveDocs == nil || de.liveDocs.At(de.accum) {
		// if (DEBUG) {
		//   System.out.println("  return doc=" + accum);
		// }
		de.position = 0
		de.lastStartOffset = 0
		de.doc = de.accum
		return de.doc, nil
	} else {
		// if (DEBUG) {
		//   System.out.println("  now do nextDoc()");
		// }
		return de.NextDoc()
	}
}

// TODO: in theory we could avoid loading frq block
// when not needed, ie, use skip data to load how far to
// seek the pos pointer ... instead of having to load frq
// blocks only to sum up how many positions to skip
func (de *everythingEnum) skipPositions() (err error) {
	// Skip positions now:
	toSkip := de.posPendingCount - de.freq
	// if (DEBUG) {
	//   System.out.println("      FPR.skipPositions: toSkip=" + toSkip);
	// }

	leftInBlock := LUCENE41_BLOCK_SIZE - de.posBufferUpto
	if toSkip < leftInBlock {
		end := de.posBufferUpto + toSkip
		for de.posBufferUpto < end {
			if de.indexHasPayloads {
				de.payloadByteUpto += int(de.payloadLengthBuffer[de.posBufferUpto])
			}
			de.posBufferUpto++
		}
		// if (DEBUG) {
		//   System.out.println("        skip w/in block to posBufferUpto=" + posBufferUpto);
		// }
	} else {
		toSkip -= leftInBlock
		for toSkip >= LUCENE41_BLOCK_SIZE {
			panic("niy")
		}
		de.refillPositions()
		de.payloadByteUpto = 0
		de.posBufferUpto = 0
		for de.posBufferUpto < toSkip {
			if de.indexHasPayloads {
				de.payloadByteUpto += int(de.payloadLengthBuffer[de.posBufferUpto])
			}
			de.posBufferUpto++
		}
		// if (DEBUG) {
		//   System.out.println("        skip w/in block to posBufferUpto=" + posBufferUpto);
		// }
	}

	de.position = 0
	de.lastStartOffset = 0
	return nil
}

func (de *everythingEnum) NextPosition() (pos int, err error) {
	// if (DEBUG) {
	//   System.out.println("    FPR.nextPosition posPendingCount=" + posPendingCount + " posBufferUpto=" + posBufferUpto + " payloadByteUpto=" + payloadByteUpto)// ;
	// }
	if de.posPendingFP != -1 {
		// if (DEBUG) {
		//   System.out.println("      seek pos to pendingFP=" + posPendingFP);
		// }
		if err = de.posIn.Seek(de.posPendingFP); err != nil {
			return
		}
		de.posPendingFP = -1

		if de.payPendingFP != -1 {
			// if (DEBUG) {
			//   System.out.println("      seek pay to pendingFP=" + payPendingFP);
			// }
			if err = de.payIn.Seek(de.payPendingFP); err != nil {
				return
			}
			de.payPendingFP = -1
		}

		// Force buffer refill:
		de.posBufferUpto = LUCENE41_BLOCK_SIZE
	}

	if de.posPendingCount > de.freq {
		de.skipPositions()
		de.posPendingCount = de.freq
	}

	if de.posBufferUpto == LUCENE41_BLOCK_SIZE {
		de.refillPositions()
		de.posBufferUpto = 0
	}
	de.position += int(de.posDeltaBuffer[de.posBufferUpto])

	if de.indexHasPayloads {
		de.payloadLength = int(de.payloadLengthBuffer[de.posBufferUpto])
		de.payload.Bytes = de.payloadBytes
		de.payload.Offset = de.payloadByteUpto
		de.payload.Length = de.payloadLength
		de.payloadByteUpto += de.payloadLength
	}

	if de.indexHasOffsets {
		de.startOffset = de.lastStartOffset + int(de.offsetStartDeltaBuffer[de.posBufferUpto])
		de.endOffset = de.startOffset + int(de.offsetLengthBuffer[de.posBufferUpto])
		de.lastStartOffset = de.startOffset
	}

	de.posBufferUpto++
	de.posPendingCount--
	// if (DEBUG) {
	//   System.out.println("      return pos=" + position);
	// }
	return de.position, nil
}

func (de *everythingEnum) StartOffset() (int, error) {
	return de.startOffset, nil
}

func (de *everythingEnum) EndOffset() (int, error) {
	return de.endOffset, nil
}

func (de *everythingEnum) Payload() (*util.BytesRef, error) {
	// if (DEBUG) {
	//   System.out.println("    FPR.getPayload payloadLength=" + payloadLength + " payloadByteUpto=" + payloadByteUpto);
	// }
	if de.payloadLength == 0 {
		return nil, nil
	} else {
		return de.payload, nil
	}
}

func (de *everythingEnum) Cost() int64 {
	return int64(de.docFreq)
}
