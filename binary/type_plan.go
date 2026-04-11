// Copyright 2024 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bin

import (
	"encoding/binary"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

// fieldFastDecode is a per-field decode closure that bypasses the generic
// type-switch in decodeBin/Borsh/CompactU16 for known primitive kinds. It
// writes directly into the struct field's memory via unsafe pointer arithmetic
// and skips option construction. Returns nil if the field is not eligible.
type fieldFastDecode func(d *Decoder, fv reflect.Value) error

// fieldFastEncode is the encode-side counterpart to fieldFastDecode.
type fieldFastEncode func(e *Encoder, fv reflect.Value) error

// fieldPlan caches per-struct-field reflect work that previously ran on every
// decode/encode call: tag parsing, sizeOf wiring, and BinaryUnmarshaler /
// BinaryMarshaler interface satisfaction.
//
// The plan is built lazily once per reflect.Type and cached in typePlanCache.
type fieldPlan struct {
	index           int      // index for rv.Field(i) / rt.Field(i)
	name            string   // for error messages and trace logs
	tag             fieldTag // parsed once at plan-build time
	skip            bool
	binaryExtension bool
	canInterface    bool         // PkgPath == "" (exported)
	fieldType       reflect.Type // cached field type

	// sizeOf wiring (struct-local; resolved at plan-build time):
	//
	//   sizeOfTargetIdx: this field has a `sizeof=X` tag — its decoded value
	//     is the slice length for the field at that index. -1 if not a source.
	//   sizeFromIdx:     this field's slice length is supplied by another
	//     field's value (i.e. some earlier field had `sizeof=<this name>`).
	//     -1 if not a target.
	sizeOfTargetIdx int
	sizeFromIdx     int

	// BinaryUnmarshaler / BinaryMarshaler satisfaction (Borsh decoder fast
	// path inspects these to bypass the indirect() call entirely).
	ptrImplementsUnmarshaler bool
	valImplementsUnmarshaler bool
	ptrImplementsMarshaler   bool
	valImplementsMarshaler   bool

	// Fast per-field dispatch closures populated for primitive kinds. nil
	// means the field must go through the generic decodeXxx/encodeXxx path.
	// borshFastDecode / binFastDecode are populated separately because
	// borsh hard-codes little-endian whereas bin/compact-u16 honour the
	// per-field byte-order tag.
	borshFastDecode fieldFastDecode
	borshFastEncode fieldFastEncode
	binFastDecode   fieldFastDecode
	binFastEncode   fieldFastEncode
}

// typePlan is the cached layout for a struct type. It is immutable once
// constructed and can be shared across goroutines.
type typePlan struct {
	// isComplexEnum is true when the first field has type BorshEnum and is
	// tagged with `borsh_enum` — these structs decode via deserializeComplexEnum
	// instead of the field loop.
	isComplexEnum bool

	fields []fieldPlan

	// hasSizeOf is true if any field carries a `sizeof=` tag and at least one
	// other field is its target. Enables the per-call sizes-array allocation.
	hasSizeOf bool

	// hasBinaryExtension is true if any field is tagged with `binary_extension`.
	hasBinaryExtension bool
}

var typePlanCache sync.Map // map[reflect.Type]*typePlan

// planForStruct returns the cached typePlan for the given struct type, building
// it on first sight. rt MUST be a struct.
func planForStruct(rt reflect.Type) *typePlan {
	if v, ok := typePlanCache.Load(rt); ok {
		return v.(*typePlan)
	}
	plan := buildStructPlan(rt)
	actual, _ := typePlanCache.LoadOrStore(rt, plan)
	return actual.(*typePlan)
}

func buildStructPlan(rt reflect.Type) *typePlan {
	plan := &typePlan{}
	n := rt.NumField()
	if n == 0 {
		return plan
	}

	// Detect Borsh complex enum at plan time so the hot path can short-circuit.
	first := rt.Field(0)
	if isTypeBorshEnum(first.Type) && parseFieldTag(first.Tag).IsBorshEnum {
		plan.isComplexEnum = true
		return plan
	}

	// Build a name → index map so we can resolve sizeof wiring statically.
	// This map only lives during plan construction.
	nameToIdx := make(map[string]int, n)
	for i := 0; i < n; i++ {
		nameToIdx[rt.Field(i).Name] = i
	}

	plan.fields = make([]fieldPlan, n)
	for i := 0; i < n; i++ {
		sf := rt.Field(i)
		tag := parseFieldTag(sf.Tag)

		fp := fieldPlan{
			index:           i,
			name:            sf.Name,
			tag:             *tag,
			skip:            tag.Skip,
			binaryExtension: tag.BinaryExtension,
			canInterface:    sf.PkgPath == "",
			fieldType:       sf.Type,
			sizeOfTargetIdx: -1,
			sizeFromIdx:     -1,
		}

		if tag.SizeOf != "" {
			if idx, ok := nameToIdx[tag.SizeOf]; ok {
				fp.sizeOfTargetIdx = idx
				plan.hasSizeOf = true
			}
		}
		if tag.BinaryExtension {
			plan.hasBinaryExtension = true
		}

		// Cache BinaryUnmarshaler / BinaryMarshaler interface satisfaction.
		// Used by the encoder/decoder per-field fast paths to skip the
		// indirect() call and the per-field rv.Interface() boxing.
		ptrType := reflect.PtrTo(sf.Type)
		fp.ptrImplementsUnmarshaler = ptrType.Implements(unmarshalableType)
		fp.valImplementsUnmarshaler = sf.Type.Implements(unmarshalableType)
		fp.ptrImplementsMarshaler = ptrType.Implements(marshalableType)
		fp.valImplementsMarshaler = sf.Type.Implements(marshalableType)

		// Populate fast dispatch closures for primitive kinds. Eligible only
		// when the field has no special tags that change the wire format
		// (Option/COption/sizeOf/etc.) and does not implement a custom marshal
		// interface — both would require the generic dispatch path.
		fp.assignFastClosures()

		plan.fields[i] = fp
	}

	// Second pass: wire sizeFromIdx (target ← source) now that targets are known.
	for srcIdx := range plan.fields {
		tgt := plan.fields[srcIdx].sizeOfTargetIdx
		if tgt >= 0 {
			plan.fields[tgt].sizeFromIdx = srcIdx
		}
	}

	return plan
}

// sizesScratch is a stack-allocated buffer the decode/encode hot paths can
// use to track per-field slice sizes when a struct has sizeof wiring, without
// allocating a heap map. Most Solana structs have far fewer than 16 fields.
const sizesScratchLen = 16

type sizesScratch [sizesScratchLen]int

// assignFastClosures populates the per-field fast dispatch closures when the
// field is eligible for the unsafe-write fast path. A field is eligible iff:
//
//   - It is exported (canInterface).
//   - It has no Option/COption/Skip/BinaryExtension/SizeOf wiring (those
//     require running the generic dispatch path).
//   - It does not implement BinaryMarshaler/BinaryUnmarshaler (custom logic).
//   - Its kind is one of the supported primitives.
//
// The closures use unsafe pointer writes against `fv.UnsafeAddr()` so they
// work uniformly for both `uint64` fields and named-type aliases (`type Slot
// uint64`) — the in-memory layout is identical when the kind matches.
func (fp *fieldPlan) assignFastClosures() {
	if !fp.canInterface ||
		fp.skip ||
		fp.binaryExtension ||
		fp.tag.Option ||
		fp.tag.COption ||
		fp.tag.SizeOf != "" ||
		fp.sizeOfTargetIdx >= 0 ||
		fp.sizeFromIdx >= 0 ||
		fp.ptrImplementsUnmarshaler ||
		fp.valImplementsUnmarshaler ||
		fp.ptrImplementsMarshaler ||
		fp.valImplementsMarshaler {
		return
	}

	// Borsh always uses little-endian. The bin/compact-u16 paths read the
	// field's tag-specified byte order, captured here so the closure doesn't
	// need to consult any per-call option.
	binOrder := fp.tag.Order
	if binOrder == nil {
		binOrder = defaultByteOrder
	}

	switch fp.fieldType.Kind() {
	case reflect.Uint8:
		fp.borshFastDecode = fastDecodeUint8
		fp.borshFastEncode = fastEncodeUint8
		fp.binFastDecode = fastDecodeUint8
		fp.binFastEncode = fastEncodeUint8
	case reflect.Int8:
		fp.borshFastDecode = fastDecodeInt8
		fp.borshFastEncode = fastEncodeInt8
		fp.binFastDecode = fastDecodeInt8
		fp.binFastEncode = fastEncodeInt8
	case reflect.Bool:
		fp.borshFastDecode = fastDecodeBool
		fp.borshFastEncode = fastEncodeBool
		fp.binFastDecode = fastDecodeBool
		fp.binFastEncode = fastEncodeBool
	case reflect.Uint16:
		fp.borshFastDecode = fastDecodeUint16LE
		fp.borshFastEncode = fastEncodeUint16LE
		fp.binFastDecode = makeBinFastDecodeUint16(binOrder)
		fp.binFastEncode = makeBinFastEncodeUint16(binOrder)
	case reflect.Int16:
		fp.borshFastDecode = fastDecodeInt16LE
		fp.borshFastEncode = fastEncodeInt16LE
		fp.binFastDecode = makeBinFastDecodeInt16(binOrder)
		fp.binFastEncode = makeBinFastEncodeInt16(binOrder)
	case reflect.Uint32:
		fp.borshFastDecode = fastDecodeUint32LE
		fp.borshFastEncode = fastEncodeUint32LE
		fp.binFastDecode = makeBinFastDecodeUint32(binOrder)
		fp.binFastEncode = makeBinFastEncodeUint32(binOrder)
	case reflect.Int32:
		fp.borshFastDecode = fastDecodeInt32LE
		fp.borshFastEncode = fastEncodeInt32LE
		fp.binFastDecode = makeBinFastDecodeInt32(binOrder)
		fp.binFastEncode = makeBinFastEncodeInt32(binOrder)
	case reflect.Uint64:
		fp.borshFastDecode = fastDecodeUint64LE
		fp.borshFastEncode = fastEncodeUint64LE
		fp.binFastDecode = makeBinFastDecodeUint64(binOrder)
		fp.binFastEncode = makeBinFastEncodeUint64(binOrder)
	case reflect.Int64:
		fp.borshFastDecode = fastDecodeInt64LE
		fp.borshFastEncode = fastEncodeInt64LE
		fp.binFastDecode = makeBinFastDecodeInt64(binOrder)
		fp.binFastEncode = makeBinFastEncodeInt64(binOrder)
	}
}

// ---- Fast primitive closures (Borsh + LE bin/compact-u16) ----
//
// All decoders write straight into the destination via fv.UnsafeAddr(); all
// encoders read straight out the same way. fv.UnsafeAddr() is safe here
// because plan-driven hot paths only call these on fields obtained via
// rv.Field(i) on a top-level Decode/Encode value (always addressable).

func fastDecodeUint8(d *Decoder, fv reflect.Value) error {
	if d.pos >= len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*uint8)(unsafe.Pointer(fv.UnsafeAddr())) = d.data[d.pos]
	d.pos++
	return nil
}

func fastEncodeUint8(e *Encoder, fv reflect.Value) error {
	return e.WriteByte(*(*uint8)(unsafe.Pointer(fv.UnsafeAddr())))
}

func fastDecodeInt8(d *Decoder, fv reflect.Value) error {
	if d.pos >= len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*int8)(unsafe.Pointer(fv.UnsafeAddr())) = int8(d.data[d.pos])
	d.pos++
	return nil
}

func fastEncodeInt8(e *Encoder, fv reflect.Value) error {
	return e.WriteByte(byte(*(*int8)(unsafe.Pointer(fv.UnsafeAddr()))))
}

func fastDecodeBool(d *Decoder, fv reflect.Value) error {
	if d.pos >= len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*bool)(unsafe.Pointer(fv.UnsafeAddr())) = d.data[d.pos] != 0
	d.pos++
	return nil
}

func fastEncodeBool(e *Encoder, fv reflect.Value) error {
	if *(*bool)(unsafe.Pointer(fv.UnsafeAddr())) {
		return e.WriteByte(1)
	}
	return e.WriteByte(0)
}

func fastDecodeUint16LE(d *Decoder, fv reflect.Value) error {
	if d.pos+2 > len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*uint16)(unsafe.Pointer(fv.UnsafeAddr())) = binary.LittleEndian.Uint16(d.data[d.pos:])
	d.pos += 2
	return nil
}

func fastEncodeUint16LE(e *Encoder, fv reflect.Value) error {
	return e.WriteUint16(*(*uint16)(unsafe.Pointer(fv.UnsafeAddr())), binary.LittleEndian)
}

func fastDecodeInt16LE(d *Decoder, fv reflect.Value) error {
	if d.pos+2 > len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*int16)(unsafe.Pointer(fv.UnsafeAddr())) = int16(binary.LittleEndian.Uint16(d.data[d.pos:]))
	d.pos += 2
	return nil
}

func fastEncodeInt16LE(e *Encoder, fv reflect.Value) error {
	return e.WriteInt16(*(*int16)(unsafe.Pointer(fv.UnsafeAddr())), binary.LittleEndian)
}

func fastDecodeUint32LE(d *Decoder, fv reflect.Value) error {
	if d.pos+4 > len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*uint32)(unsafe.Pointer(fv.UnsafeAddr())) = binary.LittleEndian.Uint32(d.data[d.pos:])
	d.pos += 4
	return nil
}

func fastEncodeUint32LE(e *Encoder, fv reflect.Value) error {
	return e.WriteUint32(*(*uint32)(unsafe.Pointer(fv.UnsafeAddr())), binary.LittleEndian)
}

func fastDecodeInt32LE(d *Decoder, fv reflect.Value) error {
	if d.pos+4 > len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*int32)(unsafe.Pointer(fv.UnsafeAddr())) = int32(binary.LittleEndian.Uint32(d.data[d.pos:]))
	d.pos += 4
	return nil
}

func fastEncodeInt32LE(e *Encoder, fv reflect.Value) error {
	return e.WriteInt32(*(*int32)(unsafe.Pointer(fv.UnsafeAddr())), binary.LittleEndian)
}

func fastDecodeUint64LE(d *Decoder, fv reflect.Value) error {
	if d.pos+8 > len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*uint64)(unsafe.Pointer(fv.UnsafeAddr())) = binary.LittleEndian.Uint64(d.data[d.pos:])
	d.pos += 8
	return nil
}

func fastEncodeUint64LE(e *Encoder, fv reflect.Value) error {
	return e.WriteUint64(*(*uint64)(unsafe.Pointer(fv.UnsafeAddr())), binary.LittleEndian)
}

func fastDecodeInt64LE(d *Decoder, fv reflect.Value) error {
	if d.pos+8 > len(d.data) {
		return io.ErrUnexpectedEOF
	}
	*(*int64)(unsafe.Pointer(fv.UnsafeAddr())) = int64(binary.LittleEndian.Uint64(d.data[d.pos:]))
	d.pos += 8
	return nil
}

func fastEncodeInt64LE(e *Encoder, fv reflect.Value) error {
	return e.WriteInt64(*(*int64)(unsafe.Pointer(fv.UnsafeAddr())), binary.LittleEndian)
}

// ---- Bin/CompactU16 fast closures (configurable byte order via tag) ----

func makeBinFastDecodeUint16(order binary.ByteOrder) fieldFastDecode {
	if order == binary.LittleEndian {
		return fastDecodeUint16LE
	}
	return func(d *Decoder, fv reflect.Value) error {
		if d.pos+2 > len(d.data) {
			return io.ErrUnexpectedEOF
		}
		*(*uint16)(unsafe.Pointer(fv.UnsafeAddr())) = order.Uint16(d.data[d.pos:])
		d.pos += 2
		return nil
	}
}

func makeBinFastEncodeUint16(order binary.ByteOrder) fieldFastEncode {
	if order == binary.LittleEndian {
		return fastEncodeUint16LE
	}
	return func(e *Encoder, fv reflect.Value) error {
		return e.WriteUint16(*(*uint16)(unsafe.Pointer(fv.UnsafeAddr())), order)
	}
}

func makeBinFastDecodeInt16(order binary.ByteOrder) fieldFastDecode {
	if order == binary.LittleEndian {
		return fastDecodeInt16LE
	}
	return func(d *Decoder, fv reflect.Value) error {
		if d.pos+2 > len(d.data) {
			return io.ErrUnexpectedEOF
		}
		*(*int16)(unsafe.Pointer(fv.UnsafeAddr())) = int16(order.Uint16(d.data[d.pos:]))
		d.pos += 2
		return nil
	}
}

func makeBinFastEncodeInt16(order binary.ByteOrder) fieldFastEncode {
	if order == binary.LittleEndian {
		return fastEncodeInt16LE
	}
	return func(e *Encoder, fv reflect.Value) error {
		return e.WriteInt16(*(*int16)(unsafe.Pointer(fv.UnsafeAddr())), order)
	}
}

func makeBinFastDecodeUint32(order binary.ByteOrder) fieldFastDecode {
	if order == binary.LittleEndian {
		return fastDecodeUint32LE
	}
	return func(d *Decoder, fv reflect.Value) error {
		if d.pos+4 > len(d.data) {
			return io.ErrUnexpectedEOF
		}
		*(*uint32)(unsafe.Pointer(fv.UnsafeAddr())) = order.Uint32(d.data[d.pos:])
		d.pos += 4
		return nil
	}
}

func makeBinFastEncodeUint32(order binary.ByteOrder) fieldFastEncode {
	if order == binary.LittleEndian {
		return fastEncodeUint32LE
	}
	return func(e *Encoder, fv reflect.Value) error {
		return e.WriteUint32(*(*uint32)(unsafe.Pointer(fv.UnsafeAddr())), order)
	}
}

func makeBinFastDecodeInt32(order binary.ByteOrder) fieldFastDecode {
	if order == binary.LittleEndian {
		return fastDecodeInt32LE
	}
	return func(d *Decoder, fv reflect.Value) error {
		if d.pos+4 > len(d.data) {
			return io.ErrUnexpectedEOF
		}
		*(*int32)(unsafe.Pointer(fv.UnsafeAddr())) = int32(order.Uint32(d.data[d.pos:]))
		d.pos += 4
		return nil
	}
}

func makeBinFastEncodeInt32(order binary.ByteOrder) fieldFastEncode {
	if order == binary.LittleEndian {
		return fastEncodeInt32LE
	}
	return func(e *Encoder, fv reflect.Value) error {
		return e.WriteInt32(*(*int32)(unsafe.Pointer(fv.UnsafeAddr())), order)
	}
}

func makeBinFastDecodeUint64(order binary.ByteOrder) fieldFastDecode {
	if order == binary.LittleEndian {
		return fastDecodeUint64LE
	}
	return func(d *Decoder, fv reflect.Value) error {
		if d.pos+8 > len(d.data) {
			return io.ErrUnexpectedEOF
		}
		*(*uint64)(unsafe.Pointer(fv.UnsafeAddr())) = order.Uint64(d.data[d.pos:])
		d.pos += 8
		return nil
	}
}

func makeBinFastEncodeUint64(order binary.ByteOrder) fieldFastEncode {
	if order == binary.LittleEndian {
		return fastEncodeUint64LE
	}
	return func(e *Encoder, fv reflect.Value) error {
		return e.WriteUint64(*(*uint64)(unsafe.Pointer(fv.UnsafeAddr())), order)
	}
}

func makeBinFastDecodeInt64(order binary.ByteOrder) fieldFastDecode {
	if order == binary.LittleEndian {
		return fastDecodeInt64LE
	}
	return func(d *Decoder, fv reflect.Value) error {
		if d.pos+8 > len(d.data) {
			return io.ErrUnexpectedEOF
		}
		*(*int64)(unsafe.Pointer(fv.UnsafeAddr())) = int64(order.Uint64(d.data[d.pos:]))
		d.pos += 8
		return nil
	}
}

func makeBinFastEncodeInt64(order binary.ByteOrder) fieldFastEncode {
	if order == binary.LittleEndian {
		return fastEncodeInt64LE
	}
	return func(e *Encoder, fv reflect.Value) error {
		return e.WriteInt64(*(*int64)(unsafe.Pointer(fv.UnsafeAddr())), order)
	}
}
