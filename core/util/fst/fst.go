package fst

import (
	"bytes"
	"fmt"
	"github.com/jtejido/golucene/core/codec"
	"github.com/jtejido/golucene/core/store"
	"github.com/jtejido/golucene/core/util"
	"math/bits"
	"reflect"
)

// util/fst/FST.java
var (
	BASE_RAM_BYTES_USED        = util.ShallowSizeOfInstance(reflect.TypeOf(FST{}))
	ARC_SHALLOW_RAM_BYTES_USED = util.ShallowSizeOfInstance(reflect.TypeOf(Arc{}))
)

type InputType int

const (
	INPUT_TYPE_BYTE1 = 1
	INPUT_TYPE_BYTE2 = 2
	INPUT_TYPE_BYTE4 = 3
)

const (
	maxInt                       = 1<<(bits.UintSize-1) - 1
	minInt                       = -maxInt - 1
	FST_BIT_FINAL_ARC            = byte(1 << 0)
	FST_BIT_LAST_ARC             = byte(1 << 1)
	FST_BIT_TARGET_NEXT          = byte(1 << 2)
	FST_BIT_STOP_NODE            = byte(1 << 3)
	FST_BIT_ARC_HAS_OUTPUT       = byte(1 << 4)
	FST_BIT_ARC_HAS_FINAL_OUTPUT = byte(1 << 5)
	FST_ARCS_AS_ARRAY_PACKED     = FST_BIT_ARC_HAS_FINAL_OUTPUT

	FST_BIT_MISSING_ARC         = byte(1 << 6)
	FST_ARCS_AS_ARRAY_WITH_GAPS = FST_BIT_MISSING_ARC
	DEFAULT_MAX_BLOCK_BITS      = 28

	FIXED_ARRAY_SHALLOW_DISTANCE = 3 // 0 => only root node
	FIXED_ARRAY_NUM_ARCS_SHALLOW = 5
	FIXED_ARRAY_NUM_ARCS_DEEP    = 10

	FST_FILE_FORMAT_NAME    = "FST"
	FST_VERSION_START       = 3
	FST_VERSION_VINT_TARGET = 4

	VERSION_CURRENT = FST_VERSION_VINT_TARGET

	FST_FINAL_END_NODE     = -1
	FST_NON_FINAL_END_NODE = 0

	/** If arc has this label then that arc is final/accepted */
	FST_END_LABEL = -1
)

// Represents a single arc
type Arc struct {
	Label           int
	Output          interface{}
	target          int64 // to node
	flags           byte
	NextFinalOutput interface{}
	nextArc         int64
	posArcsStart    int64
	bytesPerArc     int
	arcIdx          int
	numArcs         int
}

func (arc *Arc) copyFrom(other *Arc) *Arc {
	arc.Label = other.Label
	arc.target = other.target
	arc.flags = other.flags
	arc.Output = other.Output
	arc.NextFinalOutput = other.NextFinalOutput
	arc.nextArc = other.nextArc
	arc.bytesPerArc = other.bytesPerArc
	if other.bytesPerArc != 0 {
		arc.posArcsStart = other.posArcsStart
		arc.arcIdx = other.arcIdx
		arc.numArcs = other.numArcs
	}
	return arc
}

func (arc *Arc) flag(flag byte) bool {
	return hasFlag(arc.flags, flag)
}

func (arc *Arc) isLast() bool {
	return arc.flag(FST_BIT_LAST_ARC)
}

func (arc *Arc) IsFinal() bool {
	return arc.flag(FST_BIT_FINAL_ARC)
}
func (arc *Arc) isPackedArray() bool {
	return arc.bytesPerArc != 0 && arc.arcIdx > minInt
}

func (arc *Arc) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "target=%v label=%v", arc.target, util.ItoHex(int64(arc.Label)))
	if arc.flag(FST_BIT_FINAL_ARC) {
		fmt.Fprintf(&b, " final")
	}
	if arc.flag(FST_BIT_LAST_ARC) {
		fmt.Fprintf(&b, " last")
	}
	if arc.flag(FST_BIT_TARGET_NEXT) {
		fmt.Fprintf(&b, " targetNext")
	}
	if arc.flag(FST_BIT_STOP_NODE) {
		fmt.Fprintf(&b, " stop")
	}
	if arc.flag(FST_BIT_ARC_HAS_OUTPUT) {
		fmt.Fprintf(&b, " output=%v", arc.Output)
	}
	if arc.flag(FST_BIT_ARC_HAS_FINAL_OUTPUT) {
		fmt.Fprintf(&b, " nextFinalOutput=%v", arc.NextFinalOutput)
	}
	if arc.bytesPerArc != 0 {
		fmt.Fprintf(&b, " arcArray(idx=%v of %v)", arc.arcIdx, arc.numArcs)
	}
	return b.String()
}

func hasFlag(flags, bit byte) bool {
	return (flags & bit) != 0
}

type FST struct {
	inputType   InputType
	bytesPerArc []int
	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput interface{}

	bytes *BytesStore

	startNode int64

	Outputs Outputs

	NO_OUTPUT interface{}
	fstStore  FSTStore

	cachedRootArcs []*Arc

	version int32

	cachedArcsBytesUsed int
	bytesArray          []byte
}

/* Make a new empty FST, for building; Builder invokes this ctor */
func newFST(inputType InputType, outputs Outputs, bytesPageBits int) *FST {
	bytes := newBytesStoreFromBits(uint32(bytesPageBits))
	// pad: ensure no node gets address 0 which is reserved to mean
	// the stop state w/ no arcs
	bytes.WriteByte(0)
	ans := &FST{
		inputType: inputType,
		Outputs:   outputs,
		version:   VERSION_CURRENT,
		bytes:     bytes,
		NO_OUTPUT: outputs.NoOutput(),
		startNode: -1,
	}

	return ans
}

func LoadFST(in util.DataInput, outputs Outputs) (fst *FST, err error) {
	return loadFST3(in, outputs, newOnHeapFSTStore(DEFAULT_MAX_BLOCK_BITS))
}

/** Load a previously saved FST; maxBlockBits allows you to
 *  control the size of the byte[] pages used to hold the FST bytes. */
func loadFST3(in util.DataInput, outputs Outputs, fstStore FSTStore) (fst *FST, err error) {

	fst = &FST{bytes: nil, fstStore: fstStore, Outputs: outputs, startNode: -1}

	// NOTE: only reads most recent format; we don't have
	// back-compat promise for FSTs (they are experimental):
	fst.version, err = codec.CheckHeader(in, FST_FILE_FORMAT_NAME, FST_VERSION_START, VERSION_CURRENT)
	if err != nil {
		return nil, err
	}

	if b, err := in.ReadByte(); err == nil {
		if b == 1 {
			// accepts empty string
			// 1 KB blocks:
			emptyBytes := newBytesStoreFromBits(10)
			if numBytes, err := in.ReadVInt(); err == nil {
				// log.Printf("Number of bytes: %v", numBytes)
				emptyBytes.CopyBytes(in, int64(numBytes))

				// log.Printf("Reverse reader.")
				reader := emptyBytes.reverseReader()
				// NoOutputs uses 0 bytes when writing its output,
				// so we have to check here else BytesStore gets
				// angry:
				if numBytes > 0 {
					reader.setPosition(int64(numBytes - 1))
				}

				// log.Printf("Reading final output from %v to %v...\n", reader, outputs)
				fst.emptyOutput, err = outputs.ReadFinalOutput(reader)
			}
		}
	}
	if err != nil {
		return nil, err
	}

	if t, err := in.ReadByte(); err == nil {
		switch t {
		case 0:
			fst.inputType = INPUT_TYPE_BYTE1
		case 1:
			fst.inputType = INPUT_TYPE_BYTE2
		case 2:
			fst.inputType = INPUT_TYPE_BYTE4
		default:
			panic(fmt.Sprintf("invalid input type %v", t))
		}
	} else {
		return nil, err
	}

	if fst.startNode, err = in.ReadVLong(); err == nil {

		if numBytes, err := in.ReadVLong(); err == nil {
			err = fst.fstStore.Init(in, numBytes)
		}
	}

	return fst, err
}

func (t *FST) ramBytesUsed(arcs []*Arc) int64 {
	var size int64
	if arcs != nil {
		size += util.ShallowSizeOf(arcs)
		for _, arc := range arcs {
			if arc != nil {
				size += ARC_SHALLOW_RAM_BYTES_USED
				if arc.Output != nil && arc.Output != t.Outputs.NoOutput() {
					size += t.Outputs.ramBytesUsed(arc.Output)
				}
				if arc.NextFinalOutput != nil && arc.NextFinalOutput != t.Outputs.NoOutput() {
					size += t.Outputs.ramBytesUsed(arc.NextFinalOutput)
				}
			}
		}
	}
	return size
}

func (t *FST) RamBytesUsed() int64 {
	var size int64 = BASE_RAM_BYTES_USED
	if t.bytesArray != nil {
		size += int64(len(t.bytesArray))
	} else {
		size += t.bytes.RamBytesUsed()
	}
	size += int64(t.cachedArcsBytesUsed)
	return size
}

func (t *FST) finish(newStartNode int64) error {
	assert(newStartNode <= t.bytes.position())

	if t.startNode != -1 {
		return fmt.Errorf("already finished")
	}

	if newStartNode == FST_FINAL_END_NODE && t.emptyOutput != nil {
		newStartNode = 0
	}

	t.startNode = newStartNode
	t.bytes.finish()
	return t.cacheRootArcs()

}

func (t *FST) cacheRootArcs() error {
	assert(t.cachedArcsBytesUsed == 0)
	var arcs []*Arc
	var count int
	arc := &Arc{}
	t.FirstArc(arc)
	if targetHasArcs(arc) {
		in := t.BytesReader()
		arcs = make([]*Arc, 0x80)
		if _, err := t.readFirstRealTargetArc(arc.target, arc, in); err != nil {
			return err
		}

		for {
			assert(arc.Label != FST_END_LABEL)
			if arc.Label < len(arcs) {
				arcs[arc.Label] = (&Arc{}).copyFrom(arc)
			} else {
				break
			}
			if arc.isLast() {
				break
			}

			if _, err := t.readNextRealArc(arc, in); err != nil {
				return err
			}
			count++
		}
	}

	cacheRAM := t.ramBytesUsed(arcs)

	// Don't cache if there are only a few arcs or if the cache would use > 20% RAM of the FST itself:
	if count >= FIXED_ARRAY_NUM_ARCS_SHALLOW && cacheRAM < t.RamBytesUsed()/5 {
		t.cachedRootArcs = arcs
		t.cachedArcsBytesUsed = int(cacheRAM)
	}

	return nil
}

func (t *FST) EmptyOutput() interface{} {
	return t.emptyOutput
}

func (t *FST) setEmptyOutput(v interface{}) {
	if t.emptyOutput != nil {
		t.emptyOutput = t.Outputs.merge(t.emptyOutput, v)
	} else {
		t.emptyOutput = v
	}
}

func (t *FST) Save(out util.DataOutput) (err error) {
	assert2(t.startNode != -1, "call finish first")

	err = codec.WriteHeader(out, FST_FILE_FORMAT_NAME, VERSION_CURRENT)

	// TODO: really we should encode this as an arc, arriving
	// to the root node, instead of special casing here:
	if err == nil {
		if t.emptyOutput != nil {
			// accepts empty string
			err = out.WriteByte(1)

			if err == nil {
				// serialize empty-string output:
				ros := store.NewRAMOutputStreamBuffer()
				err = t.Outputs.writeFinalOutput(t.emptyOutput, ros)

				if err == nil {
					emptyOutputBytes := make([]byte, ros.FilePointer())
					err = ros.WriteToBytes(emptyOutputBytes)

					length := len(emptyOutputBytes)
					if err == nil {
						// reverse
						stopAt := length / 2
						for upto := 0; upto < stopAt; upto++ {
							emptyOutputBytes[upto], emptyOutputBytes[length-upto-1] =
								emptyOutputBytes[length-upto-1], emptyOutputBytes[upto]
						}
						err = out.WriteVInt(int32(length))
						if err == nil {
							err = out.WriteBytes(emptyOutputBytes)
						}
					}

				}
			}
		} else {
			err = out.WriteByte(0)
		}
	} else {
		return err
	}

	var tb byte
	switch int(t.inputType) {
	case INPUT_TYPE_BYTE1:
		tb = 0
	case INPUT_TYPE_BYTE2:
		tb = 1
	default:
		tb = 2
	}

	if err = out.WriteByte(tb); err == nil {
		if err = out.WriteVLong(t.startNode); err == nil {
			if t.bytes != nil {

				if err = out.WriteVLong(t.bytes.position()); err == nil {
					err = t.bytes.writeTo(out)
				}

			} else {
				assert(t.fstStore != nil)
				err = t.fstStore.WriteTo(out)

			}
		}
	}
	return err
}

/**
 * Writes an automaton to a file.
 */
// public void save(final Path path) throws IOException {
//   try (OutputStream os = new BufferedOutputStream(Files.newOutputStream(path))) {
//     save(new OutputStreamDataOutput(os));
//   }
// }

/**
 * Reads an automaton from a file.
 */
// public static <T> FST<T> read(Path path, Outputs<T> outputs) throws IOException {
//   try (InputStream is = Files.newInputStream(path)) {
//     return new FST<>(new InputStreamDataInput(new BufferedInputStream(is)), outputs);
//   }
// }
func (t *FST) writeLabel(out util.DataOutput, v int) error {
	assert2(v >= 0, "v=%v", v)
	if t.inputType == INPUT_TYPE_BYTE1 {
		assert2(v <= 255, "v=%v", v)
		return out.WriteByte(byte(v))
	} else if t.inputType == INPUT_TYPE_BYTE2 {
		assert2(v <= 65535, "v=%v", v)
		return out.WriteShort(int16(v))
	} else {
		return out.WriteVInt(int32(v))
	}

}

func (t *FST) readLabel(in util.DataInput) (v int, err error) {
	switch t.inputType {
	case INPUT_TYPE_BYTE1: // Unsigned byte
		if b, err := in.ReadByte(); err == nil {
			v = int(b)
		}
	case INPUT_TYPE_BYTE2: // Unsigned short
		if s, err := in.ReadShort(); err == nil {
			v = int(s)
		}
	default:
		v, err = AsInt(in.ReadVInt())
	}
	return v, err
}

func targetHasArcs(arc *Arc) bool {
	return arc.target > 0
}

/* Serializes new node by appending its bytes to the end of the current []byte */
func (t *FST) addNode(builder *Builder, nodeIn *UnCompiledNode) (int64, error) {
	t.NO_OUTPUT = t.Outputs.NoOutput()
	// fmt.Printf("FST.addNode pos=%v numArcs=%v\n", t.bytes.position(), nodeIn.NumArcs)
	if nodeIn.NumArcs == 0 {
		if nodeIn.IsFinal {
			return FST_FINAL_END_NODE, nil
		}
		return FST_NON_FINAL_END_NODE, nil
	}

	startAddress := builder.bytes.position()
	// fmt.Printf("  startAddr=%v\n", startAddress)

	doFixedArray := t.shouldExpand(builder, nodeIn)
	if doFixedArray {
		// fmt.Println("  fixedArray")
		if len(builder.reusedBytesPerArc) < nodeIn.NumArcs {
			builder.reusedBytesPerArc = make([]int, util.Oversize(nodeIn.NumArcs, 1))
		}
	}

	builder.arcCount += int64(nodeIn.NumArcs)

	lastArc := nodeIn.NumArcs - 1

	lastArcStart := builder.bytes.position()
	maxBytesPerArc := 0
	for arcIdx := 0; arcIdx < nodeIn.NumArcs; arcIdx++ {
		arc := nodeIn.Arcs[arcIdx]
		target := arc.Target.(*CompiledNode)
		flags := byte(0)
		// fmt.Printf("  arc %v label=%v -> target=%v\n", arcIdx, arc.label, target.node)

		if arcIdx == lastArc {
			flags += FST_BIT_LAST_ARC
		}

		if builder.lastFrozenNode == target.node && !doFixedArray {
			flags += FST_BIT_TARGET_NEXT
		}

		if arc.isFinal {
			flags += FST_BIT_FINAL_ARC
			if arc.nextFinalOutput != NO_OUTPUT {
				flags += FST_BIT_ARC_HAS_FINAL_OUTPUT
			}
		} else {
			assert(arc.nextFinalOutput == NO_OUTPUT)
		}

		targetHasArcs := target.node > 0

		if !targetHasArcs {
			flags += FST_BIT_STOP_NODE
		}

		if arc.output != NO_OUTPUT {
			flags += FST_BIT_ARC_HAS_OUTPUT
		}

		if err := builder.bytes.WriteByte(flags); err != nil {
			return 0, err
		}

		if err := t.writeLabel(builder.bytes, arc.label); err != nil {
			return 0, err
		}

		// fmt.Printf("  write arc: label=%c flags=%v target=%v pos=%v output=%v\n",
		// 	rune(arc.label), flags, target.node, t.bytes.position(),
		// 	t.outputs.outputToString(arc.output))

		if arc.output != NO_OUTPUT {
			if err := t.Outputs.Write(arc.output, builder.bytes); err != nil {
				return 0, err
			}

		}

		if arc.nextFinalOutput != NO_OUTPUT {
			// fmt.Println("    write final output")
			if err := t.Outputs.writeFinalOutput(arc.nextFinalOutput, builder.bytes); err != nil {
				return 0, err
			}
		}

		if targetHasArcs && (flags&FST_BIT_TARGET_NEXT) == 0 {
			assert(target.node > 0)
			// fmt.Println("    write target")
			if err := builder.bytes.WriteVLong(target.node); err != nil {
				return 0, err
			}
		}

		// just write the arcs "like normal" on first pass, but record
		// how many bytes each one took, and max byte size:
		if doFixedArray {
			builder.reusedBytesPerArc[arcIdx] = int(builder.bytes.position() - lastArcStart)
			lastArcStart = builder.bytes.position()
			if builder.reusedBytesPerArc[arcIdx] > maxBytesPerArc {
				maxBytesPerArc = builder.reusedBytesPerArc[arcIdx]
			}
		}
	}

	if doFixedArray {
		MAX_HEADER_SIZE := 11 // header(byte) + numArcs(vint) + numBytes(vint)
		assert(maxBytesPerArc > 0)
		// 2nd pass just "expands" all arcs to take up a fixed byte size
		// create the header
		labelRange := nodeIn.Arcs[nodeIn.NumArcs-1].label - nodeIn.Arcs[0].label + 1
		writeDirectly := labelRange > 0 && labelRange < DIRECT_ARC_LOAD_FACTOR*nodeIn.NumArcs

		header := make([]byte, MAX_HEADER_SIZE)
		bad := store.NewByteArrayDataOutput(header)
		if writeDirectly {
			bad.WriteByte(FST_ARCS_AS_ARRAY_WITH_GAPS)
			bad.WriteVInt(int32(labelRange))
		} else {
			bad.WriteByte(FST_ARCS_AS_ARRAY_PACKED)
			bad.WriteVInt(int32(nodeIn.NumArcs))
		}

		bad.WriteVInt(int32(maxBytesPerArc))
		headerLen := bad.Position()

		fixedArrayStart := startAddress + int64(headerLen)

		if writeDirectly {
			t.writeArrayWithGaps(builder, nodeIn, fixedArrayStart, maxBytesPerArc, labelRange)
		} else {
			t.writeArrayPacked(builder, nodeIn, fixedArrayStart, maxBytesPerArc)
		}

		// // expand the arcs in place, backwards
		// srcPos := builder.bytes.position()
		// destPos := fixedArrayStart + int64(nodeIn.NumArcs)*int64(maxBytesPerArc)
		// assert(destPos >= srcPos)
		// if destPos > srcPos {
		// 	builder.bytes.skipBytes(int(destPos - srcPos))
		// 	for arcIdx := nodeIn.NumArcs - 1; arcIdx >= 0; arcIdx-- {
		// 		destPos -= int64(maxBytesPerArc)
		// 		srcPos -= int64(builder.reusedBytesPerArc[arcIdx])
		// 		if srcPos != destPos {
		// 			assert2(destPos > srcPos,
		// 				"destPos=%v srcPos=%v arcIdx=%v maxBytesPerArc=%v bytesPerArc[arcIdx]=%v nodeIn.numArcs=%v",
		// 				destPos, srcPos, arcIdx, maxBytesPerArc, builder.reusedBytesPerArc[arcIdx], nodeIn.NumArcs)
		// 			builder.bytes.copyBytesInside(srcPos, destPos, builder.reusedBytesPerArc[arcIdx])
		// 		}
		// 	}
		// }

		// now write the header
		builder.bytes.writeBytesAt(startAddress, header[:headerLen])
	}

	thisNodeAddress := builder.bytes.position() - 1

	builder.bytes.reverse(startAddress, thisNodeAddress)

	builder.nodeCount++

	return thisNodeAddress, nil
}

func (t *FST) writeArrayPacked(builder *Builder, nodeIn *UnCompiledNode, fixedArrayStart int64, maxBytesPerArc int) (err error) {
	// expand the arcs in place, backwards
	srcPos := builder.bytes.position()
	destPos := fixedArrayStart + int64(nodeIn.NumArcs)*int64(maxBytesPerArc)
	assert(destPos >= srcPos)
	if destPos > srcPos {
		builder.bytes.skipBytes(int(destPos - srcPos))
		for arcIdx := nodeIn.NumArcs - 1; arcIdx >= 0; arcIdx-- {
			destPos -= int64(maxBytesPerArc)
			srcPos -= int64(builder.reusedBytesPerArc[arcIdx])
			//System.out.println("  repack arcIdx=" + arcIdx + " srcPos=" + srcPos + " destPos=" + destPos);
			if srcPos != destPos {
				//System.out.println("  copy len=" + builder.reusedBytesPerArc[arcIdx]);
				assert2(destPos > srcPos, "destPos=%d srcPos=%d arcIdx=%d maxBytesPerArc=%d reusedBytesPerArc[arcIdx]=%d nodeIn.numArcs=%d", destPos, srcPos, arcIdx, maxBytesPerArc, builder.reusedBytesPerArc[arcIdx], nodeIn.NumArcs)
				builder.bytes.copyBytesInside(srcPos, destPos, builder.reusedBytesPerArc[arcIdx])
			}
		}
	}

	return
}

func (t *FST) writeArrayWithGaps(builder *Builder, nodeIn *UnCompiledNode, fixedArrayStart int64, maxBytesPerArc, labelRange int) (err error) {
	// expand the arcs in place, backwards
	srcPos := builder.bytes.position()
	destPos := fixedArrayStart + int64(labelRange)*int64(maxBytesPerArc)
	// if destPos == srcPos it means all the arcs were the same length, and the array of them is *already* direct
	assert(destPos >= srcPos)
	if destPos > srcPos {
		builder.bytes.skipBytes(int(destPos - srcPos))
		arcIdx := nodeIn.NumArcs - 1
		firstLabel := nodeIn.Arcs[0].label
		nextLabel := nodeIn.Arcs[arcIdx].label
		for directArcIdx := labelRange - 1; directArcIdx >= 0; directArcIdx-- {
			destPos -= int64(maxBytesPerArc)
			if directArcIdx == nextLabel-firstLabel {
				arcLen := builder.reusedBytesPerArc[arcIdx]
				srcPos -= int64(arcLen)
				//System.out.println("  direct pack idx=" + directArcIdx + " arcIdx=" + arcIdx + " srcPos=" + srcPos + " destPos=" + destPos + " label=" + nextLabel);
				if srcPos != destPos {
					//System.out.println("  copy len=" + builder.reusedBytesPerArc[arcIdx]);
					assert2(destPos > srcPos, "destPos=%d srcPos=%d arcIdx=%d maxBytesPerArc=%d reusedBytesPerArc[arcIdx]=%d nodeIn.numArcs=%d", destPos, srcPos, arcIdx, maxBytesPerArc, builder.reusedBytesPerArc[arcIdx], nodeIn.NumArcs)
					builder.bytes.copyBytesInside(srcPos, destPos, arcLen)
					if arcIdx == 0 {
						break
					}
				}
				arcIdx--
				nextLabel = nodeIn.Arcs[arcIdx].label
			} else {
				assert(directArcIdx > arcIdx)
				// mark this as a missing arc
				builder.bytes.writeByteAt(destPos, FST_BIT_MISSING_ARC)
			}
		}
	}

	return
}

func (t *FST) FirstArc(arc *Arc) *Arc {
	t.NO_OUTPUT = t.Outputs.NoOutput()
	if t.emptyOutput != nil {
		arc.flags = FST_BIT_FINAL_ARC | FST_BIT_LAST_ARC
		arc.NextFinalOutput = t.emptyOutput
		if t.emptyOutput != NO_OUTPUT {
			arc.flags |= FST_BIT_ARC_HAS_FINAL_OUTPUT
		}
	} else {
		arc.flags = FST_BIT_LAST_ARC
		arc.NextFinalOutput = t.NO_OUTPUT
	}
	arc.Output = t.NO_OUTPUT

	// If there are no nodes, ie, the FST only accepts the
	// empty string, then startNode is 0
	arc.target = t.startNode
	return arc
}

func (t *FST) readLastTargetArc(follow, arc *Arc, in BytesReader) (*Arc, error) {

	//System.out.println("readLast");
	if !targetHasArcs(follow) {
		//System.out.println("  end node");
		assert(follow.IsFinal())
		arc.Label = FST_END_LABEL
		arc.target = FST_FINAL_END_NODE
		arc.Output = follow.NextFinalOutput
		arc.flags = FST_BIT_LAST_ARC
		return arc, nil
	}

	in.setPosition(follow.target)
	b, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	if b == FST_ARCS_AS_ARRAY_PACKED || b == FST_ARCS_AS_ARRAY_WITH_GAPS {
		// array: jump straight to end
		if arc.numArcs, err = AsInt(in.ReadVInt()); err == nil {
			if arc.bytesPerArc, err = AsInt(in.ReadVInt()); err == nil {
				arc.posArcsStart = in.getPosition()
			}
		}

		if err != nil {
			return nil, err
		}

		if b == FST_ARCS_AS_ARRAY_WITH_GAPS {
			arc.arcIdx = minInt
			arc.nextArc = arc.posArcsStart - int64(arc.numArcs-1)*int64(arc.bytesPerArc)
		} else {
			arc.arcIdx = arc.numArcs - 2
		}
	} else {
		arc.flags = b
		// non-array: linear scan
		arc.bytesPerArc = 0
		//System.out.println("  scan");
		for !arc.isLast() {
			// skip this arc:
			t.readLabel(in)
			if arc.flag(FST_BIT_ARC_HAS_OUTPUT) {
				t.Outputs.SkipOutput(in)
			}
			if arc.flag(FST_BIT_ARC_HAS_FINAL_OUTPUT) {
				t.Outputs.SkipFinalOutput(in)
			}
			if arc.flag(FST_BIT_STOP_NODE) {
			} else if arc.flag(FST_BIT_TARGET_NEXT) {
			} else {
				t.readUnpackedNodeTarget(in)
			}
			arc.flags, err = in.ReadByte()
			if err != nil {
				return nil, err
			}
		}
		// Undo the byte flags we read:
		in.skipBytes(-1)
		arc.nextArc = in.getPosition()
	}
	t.readNextRealArc(arc, in)
	assert(arc.isLast())
	return arc, nil

}

func (t *FST) readUnpackedNodeTarget(in BytesReader) (target int64, err error) {
	return in.ReadVLong()
}

func (t *FST) readFirstTargetArc(follow, arc *Arc, in BytesReader) (*Arc, error) {
	if follow.IsFinal() {
		// insert "fake" final first arc:
		arc.Label = FST_END_LABEL
		arc.Output = follow.NextFinalOutput
		arc.flags = FST_BIT_FINAL_ARC
		if follow.target <= 0 {
			arc.flags |= FST_BIT_LAST_ARC
		} else {

			arc.nextArc = follow.target
		}
		arc.target = FST_FINAL_END_NODE
		return arc, nil
	}
	return t.readFirstRealTargetArc(follow.target, arc, in)
}

func (t *FST) readFirstRealTargetArc(nodeAddress int64, arc *Arc, in BytesReader) (ans *Arc, err error) {

	in.setPosition(nodeAddress)

	flags, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	if flags == FST_ARCS_AS_ARRAY_PACKED || flags == FST_ARCS_AS_ARRAY_WITH_GAPS {
		// this is first arc in a fixed-array
		arc.numArcs, err = AsInt(in.ReadVInt())
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc, err = AsInt(in.ReadVInt())
		if err != nil {
			return nil, err
		}

		if flags == FST_ARCS_AS_ARRAY_PACKED {
			arc.arcIdx = -1
		} else {
			arc.arcIdx = minInt
		}

		pos := in.getPosition()
		arc.nextArc, arc.posArcsStart = pos, pos
	} else {
		// arc.flags = b
		arc.nextArc = nodeAddress
		arc.bytesPerArc = 0
	}

	return t.readNextRealArc(arc, in)
}

func (t *FST) isExpandedTarget(follow *Arc, in BytesReader) bool {
	if !targetHasArcs(follow) {
		return false
	}

	in.setPosition(follow.target)
	flags, _ := in.ReadByte()
	return flags == FST_ARCS_AS_ARRAY_PACKED || flags == FST_ARCS_AS_ARRAY_WITH_GAPS
}

func (t *FST) readNextArc(arc *Arc, in BytesReader) (*Arc, error) {
	if arc.Label == FST_END_LABEL {
		// this was a fake inserted "final" arc
		assert2(arc.nextArc > 0, "cannot readNextArc when arc.isLast()=true")
		return t.readFirstRealTargetArc(arc.nextArc, arc, in)
	} else {
		return t.readNextRealArc(arc, in)
	}
}

/** Peeks at next arc's label; does not alter arc.  Do
 *  not call this if arc.isLast()! */
func (t *FST) readNextArcLabel(arc *Arc, in BytesReader) (int, error) {
	assert(!arc.isLast())

	if arc.Label == FST_END_LABEL {
		//System.out.println("    nextArc fake " +
		//arc.nextArc);

		pos := arc.nextArc
		in.setPosition(pos)

		flags, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		if flags == FST_ARCS_AS_ARRAY_PACKED || flags == FST_ARCS_AS_ARRAY_WITH_GAPS {
			in.ReadVInt()

			// Skip bytesPerArc:
			in.ReadVInt()
		} else {
			in.setPosition(pos)
		}
		// skip flags
		in.ReadByte()
	} else {
		if arc.bytesPerArc != 0 {
			//System.out.println("    nextArc real array");
			// arcs are in an array
			if arc.arcIdx >= 0 {
				in.setPosition(arc.posArcsStart)
				// point at next arc, -1 to skip flags
				in.skipBytes((1+int64(arc.arcIdx))*int64(arc.bytesPerArc) + 1)
			} else {
				in.setPosition(arc.nextArc)
				flags, err := in.ReadByte()

				if err != nil {
					return 0, err
				}
				// skip missing arcs
				for hasFlag(flags, FST_BIT_MISSING_ARC) {
					in.skipBytes(int64(arc.bytesPerArc) - 1)
					flags, err = in.ReadByte()
					if err != nil {
						return 0, err
					}
				}
			}
		} else {
			// arcs are packed
			//System.out.println("    nextArc real packed");
			// -1 to skip flags
			in.setPosition(arc.nextArc - 1)
		}
	}
	return t.readLabel(in)
}

func (t *FST) readArcAtPosition(arc *Arc, in BytesReader, pos int64) (*Arc, error) {
	var err error
	in.setPosition(pos)
	arc.flags, err = in.ReadByte()
	if err != nil {
		return nil, err
	}

	arc.nextArc = pos
	for hasFlag(arc.flags, FST_BIT_MISSING_ARC) {
		// skip empty arcs
		arc.nextArc -= int64(arc.bytesPerArc)
		in.skipBytes(int64(arc.bytesPerArc) - 1)
		arc.flags, err = in.ReadByte()
		if err != nil {
			return nil, err
		}
	}
	return t.readArc(arc, in)
}

func (t *FST) readArcByIndex(arc *Arc, in BytesReader, idx int) (*Arc, error) {
	var err error
	arc.arcIdx = idx
	assert(arc.arcIdx < arc.numArcs)
	in.setPosition(arc.posArcsStart - int64(arc.arcIdx)*int64(arc.bytesPerArc))
	arc.flags, err = in.ReadByte()
	if err != nil {
		return nil, err
	}

	return t.readArc(arc, in)
}

/** Never returns null, but you should never call this if
 *  arc.isLast() is true. */
func (t *FST) readNextRealArc(arc *Arc, in BytesReader) (*Arc, error) {
	var err error
	// TODO: can't assert this because we call from readFirstArc
	// assert !flag(arc.flags, BIT_LAST_ARC);

	// this is a continuing arc in a fixed array
	if arc.bytesPerArc != 0 {
		// arcs are in an array
		if arc.arcIdx > minInt {
			arc.arcIdx++
			assert(arc.arcIdx < arc.numArcs)
			in.setPosition(arc.posArcsStart - int64(arc.arcIdx)*int64(arc.bytesPerArc))
			arc.flags, err = in.ReadByte()
			if err != nil {
				return nil, err
			}
		} else {
			assert(arc.nextArc <= arc.posArcsStart && arc.nextArc > arc.posArcsStart-int64(arc.numArcs)*int64(arc.bytesPerArc))
			in.setPosition(arc.nextArc)
			arc.flags, err = in.ReadByte()
			if err != nil {
				return nil, err
			}
			for hasFlag(arc.flags, FST_BIT_MISSING_ARC) {
				// skip empty arcs
				arc.nextArc = arc.nextArc - int64(arc.bytesPerArc)
				in.skipBytes(int64(arc.bytesPerArc) - 1)
				arc.flags, err = in.ReadByte()
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		// arcs are packed
		in.setPosition(arc.nextArc)
		arc.flags, err = in.ReadByte()

		if err != nil {
			return nil, err
		}
	}
	return t.readArc(arc, in)
}

func (t *FST) readArc(arc *Arc, in BytesReader) (ans *Arc, err error) {

	arc.Label, err = t.readLabel(in)

	if err != nil {
		return nil, err
	}

	if arc.flag(FST_BIT_ARC_HAS_OUTPUT) {
		arc.Output, err = t.Outputs.Read(in)
		if err != nil {
			return nil, err
		}
	} else {
		arc.Output = t.Outputs.NoOutput()
	}

	if arc.flag(FST_BIT_ARC_HAS_FINAL_OUTPUT) {
		arc.NextFinalOutput, err = t.Outputs.ReadFinalOutput(in)
		if err != nil {
			return nil, err
		}
	} else {
		arc.NextFinalOutput = t.Outputs.NoOutput()
	}

	if arc.flag(FST_BIT_STOP_NODE) {
		if arc.flag(FST_BIT_FINAL_ARC) {
			arc.target = FST_FINAL_END_NODE
		} else {
			arc.target = FST_NON_FINAL_END_NODE
		}
		if arc.bytesPerArc == 0 {
			arc.nextArc = in.getPosition()
		} else {
			arc.nextArc -= int64(arc.bytesPerArc)
		}
	} else if arc.flag(FST_BIT_TARGET_NEXT) {
		arc.nextArc = in.getPosition()
		// TODO: would be nice to make this lazy -- maybe
		// caller doesn't need the target and is scanning arcs...
		if !arc.flag(FST_BIT_LAST_ARC) {
			if arc.bytesPerArc == 0 {
				// must scan
				t.seekToNextNode(in)
			} else {
				in.setPosition(arc.posArcsStart)
				in.skipBytes(int64(arc.bytesPerArc * arc.numArcs))
			}
		}
		arc.target = in.getPosition()
	} else {
		arc.target, err = t.readUnpackedNodeTarget(in)
		if err != nil {
			return nil, err
		}
		if arc.bytesPerArc > 0 && arc.arcIdx == minInt {
			// nextArc was pointing to *this* arc when we entered; advance to the next
			// if it is a missing arc, we will skip it later
			arc.nextArc = arc.nextArc - int64(arc.bytesPerArc)
		} else {
			// in list and fixed table encodings, the next arc always follows this one
			arc.nextArc = in.getPosition()
		}
	}
	return arc, nil
}

func (t *FST) readEndArc(follow, arc *Arc) *Arc {
	if follow.IsFinal() {
		if follow.target <= 0 {
			arc.flags = FST_BIT_LAST_ARC
		} else {
			arc.flags = 0
			// NOTE: nextArc is a node (not an address!) in this case:
			arc.nextArc = follow.target
		}
		arc.Output = follow.NextFinalOutput
		arc.Label = FST_END_LABEL
		return arc
	}

	return nil
}

func (t *FST) assertRootCachedArc(label int, cachedArc *Arc) {
	arc := &Arc{}
	t.FirstArc(arc)
	in := t.BytesReader()
	result, err := t.findTargetArc(label, arc, arc, in, false)
	if err != nil {
		panic("assert failed")
	}
	if result == nil {
		assert(cachedArc == nil)
	} else {
		assert(cachedArc != nil)
		assert(cachedArc.arcIdx == result.arcIdx)
		assert(cachedArc.bytesPerArc == result.bytesPerArc)
		assert(cachedArc.flags == result.flags)
		assert(cachedArc.Label == result.Label)
		assert(cachedArc.nextArc == result.nextArc)
		assert(equals(cachedArc.NextFinalOutput, result.NextFinalOutput))
		assert(cachedArc.numArcs == result.numArcs)
		assert(equals(cachedArc.Output, result.Output))
		assert(cachedArc.posArcsStart == result.posArcsStart)
		assert(cachedArc.target == result.target)
	}

}

func (t *FST) FindTargetArc(labelToMatch int, follow *Arc, arc *Arc, in BytesReader) (target *Arc, err error) {
	return t.findTargetArc(labelToMatch, follow, arc, in, true)
}

// TODO: could we somehow [partially] tableize arc lookups
// look automaton?

/** Finds an arc leaving the incoming arc, replacing the arc in place.
 *  This returns null if the arc was not found, else the incoming arc. */
func (t *FST) findTargetArc(labelToMatch int, follow *Arc, arc *Arc, in BytesReader, useRootArcCache bool) (target *Arc, err error) {
	if labelToMatch == FST_END_LABEL {
		if follow.IsFinal() {
			if follow.target <= 0 {
				arc.flags = FST_BIT_LAST_ARC
			} else {
				arc.flags = 0
				// NOTE: nextArc is a node (not an address!) in this case:
				arc.nextArc = follow.target

			}
			arc.Output = follow.NextFinalOutput
			arc.Label = FST_END_LABEL
			return arc, nil
		} else {
			return nil, nil
		}
	}

	// Short-circuit if this arc is in the root arc cache:
	if useRootArcCache && t.cachedRootArcs != nil && follow.target == t.startNode && labelToMatch < len(t.cachedRootArcs) {
		result := t.cachedRootArcs[labelToMatch]
		// LUCENE-5152: detect tricky cases where caller
		// modified previously returned cached root-arcs:
		t.assertRootCachedArc(labelToMatch, result)
		if result != nil {
			arc.copyFrom(result)
			return arc, nil
		}
		return nil, nil
	}

	if !targetHasArcs(follow) {
		return nil, nil
	}

	in.setPosition(follow.target)

	// log.Printf("fta label=%v", labelToMatch)

	b, err := in.ReadByte()
	if err != nil {
		return nil, err
	}

	arc.numArcs, err = AsInt(in.ReadVInt())
	if err != nil {
		return nil, err
	}
	arc.bytesPerArc, err = AsInt(in.ReadInt())
	if err != nil {
		return nil, err
	}
	arc.posArcsStart = in.getPosition()

	if b == FST_ARCS_AS_ARRAY_WITH_GAPS {

		// Array is direct; address by label
		in.skipBytes(1)
		firstLabel, err := t.readLabel(in)
		if err != nil {
			return nil, err
		}

		arcPos := labelToMatch - firstLabel
		if arcPos == 0 {
			arc.nextArc = arc.posArcsStart
		} else if arcPos > 0 {
			if arcPos >= arc.numArcs {
				return nil, nil
			}
			in.setPosition(arc.posArcsStart - int64(arc.bytesPerArc*arcPos))
			flags, err := in.ReadByte()
			if err != nil {
				return nil, err
			}
			if hasFlag(flags, FST_BIT_MISSING_ARC) {
				return nil, nil
			}
			// point to flags that we just read
			arc.nextArc = in.getPosition() + 1
		} else {
			return nil, nil
		}
		arc.arcIdx = minInt
		return t.readNextRealArc(arc, in)
	} else if b == FST_ARCS_AS_ARRAY_PACKED {
		// Arcs are full array; do binary search:

		for low, high := 0, arc.numArcs-1; low < high; {
			// log.Println("    cycle")
			mid := int(uint(low+high) / 2)
			in.setPosition(arc.posArcsStart)
			in.skipBytes(int64(arc.bytesPerArc*mid) + 1)
			midLabel, err := t.readLabel(in)
			if err != nil {
				return nil, err
			}
			cmp := midLabel - labelToMatch
			if cmp < 0 {
				low = mid + 1
			} else if cmp > 0 {
				high = mid - 1
			} else {
				arc.arcIdx = mid - 1
				// log.Println("    found!")
				return t.readNextRealArc(arc, in)
			}
		}

		return nil, nil
	}

	// Linear scan

	if _, err = t.readFirstRealTargetArc(follow.target, arc, in); err != nil {
		return nil, err
	}

	for {
		//System.out.println("  non-bs cycle");
		// TODO: we should fix this code to not have to create
		// object for the output of every arc we scan... only
		// for the matching arc, if found
		if arc.Label == labelToMatch {
			//System.out.println("    found!");
			return arc, nil
		} else if arc.Label > labelToMatch {
			return nil, nil
		} else if arc.isLast() {
			return nil, nil
		} else {
			_, err = t.readNextRealArc(arc, in)
			if err != nil {
				return nil, err
			}
		}
	}
}

func (t *FST) seekToNextNode(in BytesReader) error {
	var err error
	var flags byte
	for {
		if flags, err = in.ReadByte(); err == nil {
			_, err = t.readLabel(in)
		}
		if err != nil {
			return err
		}

		if hasFlag(flags, FST_BIT_ARC_HAS_OUTPUT) {
			if err = t.Outputs.SkipOutput(in); err != nil {
				return err
			}
		}

		if hasFlag(flags, FST_BIT_ARC_HAS_FINAL_OUTPUT) {
			if err = t.Outputs.SkipFinalOutput(in); err != nil {
				return err
			}
		}

		if !hasFlag(flags, FST_BIT_STOP_NODE) && !hasFlag(flags, FST_BIT_TARGET_NEXT) {
			_, err = t.readUnpackedNodeTarget(in)
			if err != nil {
				return err
			}
		}

		if hasFlag(flags, FST_BIT_LAST_ARC) {
			return nil
		}
	}
}

func (t *FST) shouldExpand(builder *Builder, node *UnCompiledNode) bool {
	return builder.allowArrayArcs &&
		(node.depth <= FIXED_ARRAY_SHALLOW_DISTANCE && node.NumArcs >= FIXED_ARRAY_NUM_ARCS_SHALLOW ||
			node.NumArcs >= FIXED_ARRAY_NUM_ARCS_DEEP)
}

// Since Go doesn't has Java's Object.equals() method,
// I have to implement my own.
func equals(a, b interface{}) bool {
	sameType := reflect.TypeOf(a) == reflect.TypeOf(b)
	if _, ok := a.([]byte); ok {
		if _, ok := b.([]byte); !ok {
			// panic(fmt.Sprintf("incomparable type: %v vs %v", a, b))
			return false
		}
		b1 := a.([]byte)
		b2 := b.([]byte)
		if len(b1) != len(b2) {
			return false
		}
		for i := 0; i < len(b1) && i < len(b2); i++ {
			if b1[i] != b2[i] {
				return false
			}
		}
		return true
	} else if _, ok := a.(int64); ok {
		if _, ok := b.(int64); !ok {
			// panic(fmt.Sprintf("incomparable type: %v vs %v", a, b))
			return false
		}
		return a.(int64) == b.(int64)
	} else if a == nil && b == nil {
		return true
	} else if sameType && a == b {
		return true
	}
	return false
}

func CompareFSTValue(a, b interface{}) bool {
	return equals(a, b)
}

func AsInt(n int32, err error) (n2 int, err2 error) {
	return int(n), err
}

func AsInt64(n int32, err error) (n2 int64, err2 error) {
	return int64(n), err
}

func (t *FST) BytesReader() BytesReader {

	if t.fstStore != nil {
		return t.fstStore.ReverseBytesReader()
	}

	return t.bytes.reverseReader()
}

type RandomAccess interface {
	getPosition() int64
	setPosition(pos int64)
	reversed() bool
	skipBytes(count int64)
}

type BytesReader interface {
	// *util.DataInputImpl
	util.DataInput
	RandomAccess
}

// L1464
/*
Expert: creates an FST by packing this one. This process requires
substantial additional RAM (currently up to ~8 bytes per node
depending on acceptableOverheadRatio), but then should produce a
smaller FST.

The implementation of this method uses ideas from
<a target="_blank" href="http://www.cs.put.poznan.pl/dweiss/site/publications/download/fsacomp.pdf">Smaller Representation of Finite State Automata</a>
which describes techniques to reduce the size of a FST. However, this
is not a strict implementation of the algorithms described in this
paper.
*/
func (t *FST) pack(minInCountDeref, maxDerefNodes int,
	acceptableOverheadRatio float32) (*FST, error) {
	panic("not implemented yet")
}
