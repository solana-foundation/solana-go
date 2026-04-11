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
	"math/big"
	"reflect"
)

// isInvalidValue reports whether rv is the zero reflect.Value (Kind ==
// Invalid). It is NOT the same as rv.IsZero(): a valid reflect.Value holding
// a zero-valued T returns false here. Used by the encoders to short-circuit
// before attempting to interface() a nil reflect.Value.
func isInvalidValue(rv reflect.Value) bool {
	return rv.Kind() == reflect.Invalid
}

// asBinaryMarshaler returns a BinaryMarshaler for rv if one is reachable.
// It first tries the value itself; if that fails and rv is addressable, it
// retries via rv.Addr() so that marshalers implemented on *T are still found
// when the field is held by value. Without the second try, a legitimate
// custom marshaler is silently skipped and the encoder falls back to the
// generic reflect path — producing a different wire encoding.
//
// Performance note: rv.Interface() boxes the value into an interface{},
// which heap-allocates for any non-pointer type larger than a word. We
// short-circuit via reflect.Type.Implements (a static type-info lookup
// with no allocation) so the boxing only happens for types that actually
// satisfy BinaryMarshaler — turning the dominant per-field allocation
// site into a no-op for types like solana.PublicKey.
func asBinaryMarshaler(rv reflect.Value) (BinaryMarshaler, bool) {
	if !rv.IsValid() {
		return nil, false
	}
	rt := rv.Type()
	if rt.Implements(marshalableType) && rv.CanInterface() {
		if m, ok := rv.Interface().(BinaryMarshaler); ok {
			return m, true
		}
	}
	if rv.CanAddr() && reflect.PointerTo(rt).Implements(marshalableType) {
		addr := rv.Addr()
		if addr.CanInterface() {
			if m, ok := addr.Interface().(BinaryMarshaler); ok {
				return m, true
			}
		}
	}
	return nil, false
}

func twosComplement(v []byte) []byte {
	buf := make([]byte, len(v))
	for i, b := range v {
		buf[i] = b ^ byte(0xff)
	}
	one := big.NewInt(1)
	value := (&big.Int{}).SetBytes(buf)
	return value.Add(value, one).Bytes()
}
