package index

import (
	"github.com/balzaczyy/golucene/core/store"
)

// codec/lucene45/Lucene45Codec.java

// NOTE: if we make largish changes in a minor release, easier to
// just make Lucene46Codec or whatever if they are backwards
// compatible or smallish we can probably do the backwards in the
// postingreader (it writes a minor version, etc).
/*
Implements the Lucene 4.5 index format, with configurable per-field
postings and docvalues formats.

If you want to reuse functionality of this codec in another codec,
extend FilterCodec.
*/
var Lucene45Codec = Codec{
	ReadSegmentInfo: Lucene40SegmentInfoReader,
	ReadFieldInfos:  Lucene42FieldInfosReader,
	GetFieldsProducer: func(readState SegmentReadState) (fp FieldsProducer, err error) {
		return newPerFieldPostingsReader(readState)
	},
	GetDocValuesProducer: func(s SegmentReadState) (dvp DocValuesProducer, err error) {
		return newPerFieldDocValuesReader(s)
	},
	GetNormsDocValuesProducer: func(s SegmentReadState) (dvp DocValuesProducer, err error) {
		return newLucene42DocValuesProducer(s, "Lucene41NormsData", "nvd", "Lucene41NormsMetadata", "nvm")
	},
	GetStoredFieldsReader: func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r StoredFieldsReader, err error) {
		return newLucene41StoredFieldsReader(d, si, fn, ctx)
	},
	GetTermVectorsReader: func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error) {
		return newLucene42TermVectorsReader(d, si, fn, ctx)
	},
}

// codec/lucene45/Lucene45DocValuesFormat.java

const (
	LUCENE45_DV_DATA_CODEC     = "Lucene45DocValuesData"
	LUCENE45_DV_DATA_EXTENSION = "dvd"
	LUCENE45_DV_META_CODEC     = "Lucene45valuesMetadata"
	LUCENE45_DV_META_EXTENSION = "dvm"
)

// codec/lucene45/Lucene45DocValuesProducer.java

type Lucene45DocvaluesProducer struct {
}

// expert: instantiate a new reader
func newLucene45DocValuesProducer(
	state SegmentReadState, dataCodec, dataExtension, metaCodec, metaExtension string) (
	dvp *Lucene42DocValuesProducer, err error) {
	panic("not implemented yet")
}