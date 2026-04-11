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
	"fmt"
	"reflect"

	"go.uber.org/zap"
)

func (e *Encoder) encodeCompactU16(rv reflect.Value, opt option) (err error) {
	if opt.Order == nil {
		opt.Order = defaultByteOrder
	}
	e.currentFieldOpt = opt

	if traceEnabled {
		zlog.Debug("encode: type",
			zap.Stringer("value_kind", rv.Kind()),
			zap.Reflect("options", opt),
		)
	}

	if opt.is_Optional() {
		if rv.IsZero() {
			if traceEnabled {
				zlog.Debug("encode: skipping optional value with", zap.Stringer("type", rv.Kind()))
			}
			return e.WriteBool(false)
		}
		err := e.WriteBool(true)
		if err != nil {
			return err
		}
		// The optionality has been used; stop its propagation:
		opt.is_OptionalField = false
	}

	if isInvalidValue(rv) {
		return nil
	}

	if marshaler, ok := asBinaryMarshaler(rv); ok {
		if traceEnabled {
			zlog.Debug("encode: using MarshalerBinary method to encode type")
		}
		return marshaler.MarshalWithEncoder(e)
	}

	switch rv.Kind() {
	case reflect.String:
		return e.WriteString(rv.String())
	case reflect.Uint8:
		return e.WriteByte(byte(rv.Uint()))
	case reflect.Int8:
		return e.WriteByte(byte(rv.Int()))
	case reflect.Int16:
		return e.WriteInt16(int16(rv.Int()), opt.Order)
	case reflect.Uint16:
		return e.WriteUint16(uint16(rv.Uint()), opt.Order)
	case reflect.Int32:
		return e.WriteInt32(int32(rv.Int()), opt.Order)
	case reflect.Uint32:
		return e.WriteUint32(uint32(rv.Uint()), opt.Order)
	case reflect.Uint64:
		return e.WriteUint64(rv.Uint(), opt.Order)
	case reflect.Int64:
		return e.WriteInt64(rv.Int(), opt.Order)
	case reflect.Float32:
		return e.WriteFloat32(float32(rv.Float()), opt.Order)
	case reflect.Float64:
		return e.WriteFloat64(rv.Float(), opt.Order)
	case reflect.Bool:
		return e.WriteBool(rv.Bool())
	case reflect.Ptr:
		return e.encodeCompactU16(rv.Elem(), opt)
	case reflect.Interface:
		// skip
		return nil
	}

	rv = reflect.Indirect(rv)
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Array:
		l := rt.Len()
		if traceEnabled {
			defer func(prev *zap.Logger) { zlog = prev }(zlog)
			zlog = zlog.Named("array")
			zlog.Debug("encode: array", zap.Int("length", l), zap.Stringer("type", rv.Kind()))
		}

		switch k := rv.Type().Elem().Kind(); k {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// if it's a [n]byte, accumulate and write in one command:
			if err := reflect_writeArrayOfUint_(e, l, k, rv, LE); err != nil {
				return err
			}
		default:
			for i := range l {
				if err = e.encodeCompactU16(rv.Index(i), opt); err != nil {
					return
				}
			}
		}
	case reflect.Slice:
		var l int
		if opt.hasSizeOfSlice() {
			l = opt.getSizeOfSlice()
			if traceEnabled {
				zlog.Debug("encode: slice with sizeof set", zap.Int("size_of", l))
			}
		} else {
			l = rv.Len()
			if err = e.WriteCompactU16Length(l); err != nil {
				return
			}
		}
		if traceEnabled {
			defer func(prev *zap.Logger) { zlog = prev }(zlog)
			zlog = zlog.Named("slice")
			zlog.Debug("encode: slice", zap.Int("length", l), zap.Stringer("type", rv.Kind()))
		}

		// we would want to skip to the correct head_offset

		switch k := rv.Type().Elem().Kind(); k {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// if it's a [n]byte, accumulate and write in one command:
			if err := reflect_writeArrayOfUint_(e, l, k, rv, LE); err != nil {
				return err
			}
		default:
			for i := range l {
				if err = e.encodeCompactU16(rv.Index(i), opt); err != nil {
					return
				}
			}
		}
	case reflect.Struct:
		if err = e.encodeStructCompactU16(rt, rv); err != nil {
			return
		}

	case reflect.Map:
		keyCount := len(rv.MapKeys())

		if traceEnabled {
			zlog.Debug("encode: map",
				zap.Int("key_count", keyCount),
				zap.String("key_type", rt.String()),
				typeField("value_type", rv.Elem()),
			)
			defer func(prev *zap.Logger) { zlog = prev }(zlog)
			zlog = zlog.Named("struct")
		}

		if err = e.WriteCompactU16Length(keyCount); err != nil {
			return
		}

		for _, mapKey := range rv.MapKeys() {
			if err = e.Encode(mapKey.Interface()); err != nil {
				return
			}

			if err = e.Encode(rv.MapIndex(mapKey).Interface()); err != nil {
				return
			}
		}

	default:
		return fmt.Errorf("encode: unsupported type %q", rt)
	}
	return
}

func (e *Encoder) encodeStructCompactU16(rt reflect.Type, rv reflect.Value) (err error) {
	plan := planForStruct(rt)

	if traceEnabled {
		zlog.Debug("encode: struct", zap.Int("fields", len(plan.fields)), zap.Stringer("type", rv.Kind()))
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

	fastOK := rv.CanAddr()
	for i := range plan.fields {
		fp := &plan.fields[i]

		if fp.skip {
			if traceEnabled {
				zlog.Debug("encode: skipping struct field with skip flag",
					zap.String("struct_field_name", fp.name),
				)
			}
			continue
		}

		// Fast primitive path: no option construction, no kind switch.
		if fastOK && fp.binFastEncode != nil {
			if err := fp.binFastEncode(e, rv.Field(i)); err != nil {
				return fmt.Errorf("error while encoding %q field: %w", fp.name, err)
			}
			continue
		}

		fv := rv.Field(i)

		if fp.sizeOfTargetIdx >= 0 && sizes != nil {
			sizes[fp.sizeOfTargetIdx] = sizeof(fp.fieldType, fv)
		}

		if !fp.canInterface {
			if traceEnabled {
				zlog.Debug("encode:  skipping field: unable to interface field, probably since field is not exported",
					zap.String("struct_field_name", fp.name),
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
			zlog.Debug("encode: struct field",
				zap.Stringer("struct_field_value_type", fv.Kind()),
				zap.String("struct_field_name", fp.name),
				zap.Reflect("struct_field_tags", fp.tag),
				zap.Reflect("struct_field_option", opt),
			)
		}

		if err := e.encodeCompactU16(fv, opt); err != nil {
			return fmt.Errorf("error while encoding %q field: %w", fp.name, err)
		}
	}
	return nil
}
