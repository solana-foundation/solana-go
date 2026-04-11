// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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
	"fmt"
	"io"
	"reflect"

	"go.uber.org/zap"
)

func (dec *Decoder) decodeWithOptionBin(v interface{}, opt option) (err error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return &InvalidDecoderError{reflect.TypeOf(v)}
	}

	// We decode rv not rv.Elem because the Unmarshaler interface
	// test must be applied at the top level of the value.
	err = dec.decodeBin(rv, opt)
	if err != nil {
		return err
	}
	return nil
}

func (dec *Decoder) decodeBin(rv reflect.Value, opt option) (err error) {
	if opt.Order == nil {
		opt.Order = defaultByteOrder
	}
	dec.currentFieldOpt = opt

	unmarshaler, rv := indirect(rv, opt.is_Optional())

	if traceEnabled {
		zlog.Debug("decode: type",
			zap.Stringer("value_kind", rv.Kind()),
			zap.Bool("has_unmarshaler", (unmarshaler != nil)),
			zap.Reflect("options", opt),
		)
	}

	if opt.is_Optional() {
		isPresent, e := dec.ReadUint32(binary.LittleEndian)
		if e != nil {
			err = fmt.Errorf("decode: %s isPresent, %s", rv.Type().String(), e)
			return
		}

		if isPresent == 0 {
			if traceEnabled {
				zlog.Debug("decode: skipping optional value", zap.Stringer("type", rv.Kind()))
			}

			rv.Set(reflect.Zero(rv.Type()))
			return
		}

		// we have ptr here we should not go get the element
		unmarshaler, rv = indirect(rv, false)
	}

	if unmarshaler != nil {
		if traceEnabled {
			zlog.Debug("decode: using UnmarshalWithDecoder method to decode type")
		}
		return unmarshaler.UnmarshalWithDecoder(dec)
	}
	rt := rv.Type()

	switch rv.Kind() {
	case reflect.String:
		s, e := dec.ReadRustString()
		if e != nil {
			err = e
			return
		}
		rv.SetString(s)
		return
	case reflect.Uint8:
		var n byte
		n, err = dec.ReadByte()
		rv.SetUint(uint64(n))
		return
	case reflect.Int8:
		var n int8
		n, err = dec.ReadInt8()
		rv.SetInt(int64(n))
		return
	case reflect.Int16:
		var n int16
		n, err = dec.ReadInt16(opt.Order)
		rv.SetInt(int64(n))
		return
	case reflect.Int32:
		var n int32
		n, err = dec.ReadInt32(opt.Order)
		rv.SetInt(int64(n))
		return
	case reflect.Int64:
		var n int64
		n, err = dec.ReadInt64(opt.Order)
		rv.SetInt(int64(n))
		return
	case reflect.Uint16:
		var n uint16
		n, err = dec.ReadUint16(opt.Order)
		rv.SetUint(uint64(n))
		return
	case reflect.Uint32:
		var n uint32
		n, err = dec.ReadUint32(opt.Order)
		rv.SetUint(uint64(n))
		return
	case reflect.Uint64:
		var n uint64
		n, err = dec.ReadUint64(opt.Order)
		rv.SetUint(n)
		return
	case reflect.Float32:
		var n float32
		n, err = dec.ReadFloat32(opt.Order)
		rv.SetFloat(float64(n))
		return
	case reflect.Float64:
		var n float64
		n, err = dec.ReadFloat64(opt.Order)
		rv.SetFloat(n)
		return
	case reflect.Bool:
		var r bool
		r, err = dec.ReadBool()
		rv.SetBool(r)
		return
	case reflect.Interface:
		// skip
		return nil
	}
	switch rt.Kind() {
	case reflect.Array:
		l := rt.Len()
		if traceEnabled {
			zlog.Debug("decoding: reading array", zap.Int("length", l))
		}

		switch k := rv.Type().Elem().Kind(); k {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if err := reflect_readArrayOfUint_(dec, l, k, rv, LE); err != nil {
				return err
			}
		default:
			for i := range l {
				if err = dec.decodeBin(rv.Index(i), opt); err != nil {
					return
				}
			}
		}
		return
	case reflect.Slice:
		var l int
		if opt.hasSizeOfSlice() {
			l = opt.getSizeOfSlice()
		} else {
			length, err := dec.ReadLength()
			if err != nil {
				return err
			}
			l = length
		}

		if traceEnabled {
			zlog.Debug("reading slice", zap.Int("len", l), typeField("type", rv))
		}

		if l > dec.Remaining() {
			return io.ErrUnexpectedEOF
		}

		switch k := rv.Type().Elem().Kind(); k {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if err := reflect_readArrayOfUint_(dec, l, k, rv, LE); err != nil {
				return err
			}
		default:
			// Pre-size the slice once instead of growing it via reflect.Append
			// in a loop. Decoding into slc.Index(i).Addr() also avoids the
			// per-element reflect.New heap allocation.
			slc := reflect.MakeSlice(rt, l, l)
			elOpt := option{Order: opt.Order}
			for i := range l {
				if err = dec.decodeBin(slc.Index(i).Addr(), elOpt); err != nil {
					return
				}
			}
			rv.Set(slc)
		}

	case reflect.Struct:
		if err = dec.decodeStructBin(rt, rv); err != nil {
			return
		}

	case reflect.Map:
		l, err := dec.ReadLength()
		if err != nil {
			return err
		}
		if l == 0 {
			// If the map has no content, keep it nil.
			return nil
		}
		rv.Set(reflect.MakeMap(rt))
		mapOpt := option{Order: opt.Order}
		for i := 0; i < int(l); i++ {
			key := reflect.New(rt.Key())
			err := dec.decodeBin(key.Elem(), mapOpt)
			if err != nil {
				return err
			}
			val := reflect.New(rt.Elem())
			err = dec.decodeBin(val.Elem(), mapOpt)
			if err != nil {
				return err
			}
			rv.SetMapIndex(key.Elem(), val.Elem())
		}
		return nil

	default:
		return fmt.Errorf("decode: unsupported type %q", rt)
	}

	return
}

func (dec *Decoder) decodeStructBin(rt reflect.Type, rv reflect.Value) (err error) {
	plan := planForStruct(rt)

	if traceEnabled {
		zlog.Debug("decode: struct", zap.Int("fields", len(plan.fields)), zap.Stringer("type", rv.Kind()))
	}

	var sizes []int
	if plan.hasSizeOf {
		var stack sizesScratch
		if len(plan.fields) <= sizesScratchLen {
			sizes = stack[:len(plan.fields)]
		} else {
			sizes = make([]int, len(plan.fields))
		}
		for i := range sizes {
			sizes[i] = -1
		}
	}

	seenBinaryExtensionField := false
	for i := range plan.fields {
		fp := &plan.fields[i]

		if fp.skip {
			if traceEnabled {
				zlog.Debug("decode: skipping struct field with skip flag",
					zap.String("struct_field_name", fp.name),
				)
			}
			continue
		}

		if !fp.binaryExtension && seenBinaryExtensionField {
			panic(fmt.Sprintf("the `bin:\"binary_extension\"` tags must be packed together at the end of struct fields, problematic field %q", fp.name))
		}

		if fp.binaryExtension {
			seenBinaryExtensionField = true
			if len(dec.data[dec.pos:]) <= 0 {
				continue
			}
		}

		// Fast primitive path: no option construction, no kind switch.
		if fp.binFastDecode != nil {
			if err = fp.binFastDecode(dec, rv.Field(i)); err != nil {
				return fmt.Errorf("error while decoding %q field: %w", fp.name, err)
			}
			continue
		}

		v := rv.Field(i)
		if !v.CanSet() {
			if !v.CanAddr() {
				if traceEnabled {
					zlog.Debug("skipping struct field that cannot be addressed",
						zap.String("struct_field_name", fp.name),
						zap.Stringer("struct_value_type", v.Kind()),
					)
				}
				return fmt.Errorf("unable to decode a none setup struc field %q with type %q", fp.name, v.Kind())
			}
			v = v.Addr()
		}

		if !v.CanSet() {
			if traceEnabled {
				zlog.Debug("skipping struct field that cannot be addressed",
					zap.String("struct_field_name", fp.name),
					zap.Stringer("struct_value_type", v.Kind()),
				)
			}
			continue
		}

		opt := option{
			is_OptionalField: fp.tag.Option,
			Order:            fp.tag.Order,
		}

		if sizes != nil && fp.sizeFromIdx >= 0 && sizes[i] >= 0 {
			opt.sliceSizeIsSet = true
			opt.sliceSize = sizes[i]
		}

		if traceEnabled {
			zlog.Debug("decode: struct field",
				zap.Stringer("struct_field_value_type", v.Kind()),
				zap.String("struct_field_name", fp.name),
				zap.Reflect("struct_field_tags", fp.tag),
				zap.Reflect("struct_field_option", opt),
			)
		}

		if err = dec.decodeBin(v, opt); err != nil {
			return fmt.Errorf("error while decoding %q field: %w", fp.name, err)
		}

		if fp.sizeOfTargetIdx >= 0 && sizes != nil {
			size := sizeof(fp.fieldType, v)
			if traceEnabled {
				zlog.Debug("setting size of field",
					zap.String("field_name", plan.fields[fp.sizeOfTargetIdx].name),
					zap.Int("size", size),
				)
			}
			sizes[fp.sizeOfTargetIdx] = size
		}
	}
	return
}
