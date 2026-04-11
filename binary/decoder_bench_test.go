package bin

import (
	"reflect"
	"testing"
)

func newUint64SliceEncoded(l int) []byte {
	buf := make([]byte, 0)
	for i := 0; i < l; i++ {
		buf = append(buf, uint64ToBytes(uint64(i), LE)...)
	}
	return buf
}

func Benchmark_uintSlice64_Decode_noMake(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var got []uint64

		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}
func Benchmark_uintSlice64_Decode_make(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		got := make([]uint64, 0)

		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}

func Benchmark_uintSlice64_Decode_field_noMake(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()
	type S struct {
		Field []uint64
	}
	for i := 0; i < b.N; i++ {
		var got S

		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
		if len(got.Field) != l {
			b.Errorf("got %d, want %d", len(got.Field), l)
		}
	}
}

func Benchmark_uintSlice64_Decode_field_make(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()
	type S struct {
		Field []uint64
	}
	for i := 0; i < b.N; i++ {
		var got S
		got.Field = make([]uint64, 0)

		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
		if len(got.Field) != l {
			b.Errorf("got %d, want %d", len(got.Field), l)
		}
	}
}

func Benchmark_uintSlice64_readArray_noMake(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var got []uint64

		decoder := NewBorshDecoder(buf)
		rv := reflect.ValueOf(&got).Elem()
		k := rv.Type().Elem().Kind()

		err := reflect_readArrayOfUint_(decoder, len(buf)/8, k, rv, LE)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}

func Benchmark_uintSlice64_readArray_make(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		got := make([]uint64, 0)

		decoder := NewBorshDecoder(buf)
		rv := reflect.ValueOf(&got).Elem()
		k := rv.Type().Elem().Kind()

		err := reflect_readArrayOfUint_(decoder, len(buf)/8, k, rv, LE)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}

type sliceUint64WithCustomDecoder []uint64

// UnmarshalWithDecoder
func (s *sliceUint64WithCustomDecoder) UnmarshalWithDecoder(decoder *Decoder) error {
	// read length
	l, err := decoder.ReadUint32(LE)
	if err != nil {
		return err
	}
	// read data
	*s = make([]uint64, l)
	for i := 0; i < int(l); i++ {
		(*s)[i], err = decoder.ReadUint64(LE)
		if err != nil {
			return err
		}
	}
	return nil
}

func Benchmark_uintSlice64_Decode_field_withCustomDecoder(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var got sliceUint64WithCustomDecoder

		decoder := NewBorshDecoder(buf)
		err := got.UnmarshalWithDecoder(decoder)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}

func newUint32SliceEncoded(l int) []byte {
	buf := make([]byte, 0)
	for i := 0; i < l; i++ {
		buf = append(buf, uint32ToBytes(uint32(i), LE)...)
	}
	return buf
}

func Benchmark_uintSlice32_Decode_noMake(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint32SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var got []uint32

		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}
func Benchmark_uintSlice32_Decode_make(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint32SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		got := make([]uint32, 0)

		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}

func Benchmark_uintSlice32_Decode_field_noMake(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint32SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()
	type S struct {
		Field []uint32
	}
	for i := 0; i < b.N; i++ {
		var got S

		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
		if len(got.Field) != l {
			b.Errorf("got %d, want %d", len(got.Field), l)
		}
	}
}

func Benchmark_uintSlice32_Decode_field_make(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint32SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()
	type S struct {
		Field []uint32
	}
	for i := 0; i < b.N; i++ {
		var got S
		got.Field = make([]uint32, 0)

		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
		if len(got.Field) != l {
			b.Errorf("got %d, want %d", len(got.Field), l)
		}
	}
}

func Benchmark_uintSlice32_readArray_noMake(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		newUint32SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var got []uint32

		decoder := NewBorshDecoder(buf)
		rv := reflect.ValueOf(&got).Elem()
		k := rv.Type().Elem().Kind()

		err := reflect_readArrayOfUint_(decoder, len(buf)/4, k, rv, LE)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}

func Benchmark_uintSlice32_readArray_make(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		newUint32SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		got := make([]uint32, 0)

		decoder := NewBorshDecoder(buf)
		rv := reflect.ValueOf(&got).Elem()
		k := rv.Type().Elem().Kind()

		err := reflect_readArrayOfUint_(decoder, len(buf)/4, k, rv, LE)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}

type sliceUint32WithCustomDecoder []uint32

// UnmarshalWithDecoder
func (s *sliceUint32WithCustomDecoder) UnmarshalWithDecoder(decoder *Decoder) error {
	// read length
	l, err := decoder.ReadUint32(LE)
	if err != nil {
		return err
	}
	// read data
	*s = make([]uint32, l)
	for i := 0; i < int(l); i++ {
		(*s)[i], err = decoder.ReadUint32(LE)
		if err != nil {
			return err
		}
	}
	return nil
}
func Benchmark_uintSlice32_Decode_field_withCustomDecoder(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint32SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var got sliceUint32WithCustomDecoder

		decoder := NewBorshDecoder(buf)
		err := got.UnmarshalWithDecoder(decoder)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}
