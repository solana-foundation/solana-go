package bin

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// discardWriter is a zero-alloc io.Writer used to measure the encoder itself
// without pulling in the growth/copy costs of bytes.Buffer.
type discardWriter struct{ n int }

func (d *discardWriter) Write(p []byte) (int, error) {
	d.n += len(p)
	return len(p), nil
}

// ---- primitive writes (target of review item #1) ----

func BenchmarkEncode_WriteUint16(b *testing.B) {
	var w discardWriter
	e := NewBinEncoder(&w)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.WriteUint16(uint16(i), binary.LittleEndian)
	}
}

func BenchmarkEncode_WriteUint32(b *testing.B) {
	var w discardWriter
	e := NewBinEncoder(&w)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.WriteUint32(uint32(i), binary.LittleEndian)
	}
}

func BenchmarkEncode_WriteUint64(b *testing.B) {
	var w discardWriter
	e := NewBinEncoder(&w)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.WriteUint64(uint64(i), binary.LittleEndian)
	}
}

// review item #8
func BenchmarkEncode_WriteUint128(b *testing.B) {
	var w discardWriter
	e := NewBinEncoder(&w)
	v := Uint128{Lo: 0xdeadbeef, Hi: 0xcafebabe}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.WriteUint128(v, binary.LittleEndian)
	}
}

func BenchmarkEncode_WriteUVarInt(b *testing.B) {
	var w discardWriter
	e := NewBinEncoder(&w)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.WriteUVarInt(i)
	}
}

// ---- compact-u16 (target of review item #7) ----

func BenchmarkEncode_CompactU16_1byte(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 3)
		_ = EncodeCompactU16Length(&buf, 42)
	}
}

func BenchmarkEncode_CompactU16_2byte(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 3)
		_ = EncodeCompactU16Length(&buf, 300)
	}
}

func BenchmarkEncode_CompactU16_3byte(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 3)
		_ = EncodeCompactU16Length(&buf, 20000)
	}
}

func BenchmarkDecode_CompactU16_1byte(b *testing.B) {
	var buf []byte
	_ = EncodeCompactU16Length(&buf, 42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = DecodeCompactU16(buf)
	}
}

func BenchmarkDecode_CompactU16_3byte(b *testing.B) {
	var buf []byte
	_ = EncodeCompactU16Length(&buf, 20000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = DecodeCompactU16(buf)
	}
}

// WriteCompactU16 routes through the Encoder, which currently allocates twice
// (once for the scratch append buffer, once via toWriter).
func BenchmarkEncode_CompactU16_ViaEncoder(b *testing.B) {
	var w discardWriter
	e := NewCompactU16Encoder(&w)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.WriteCompactU16(20000)
	}
}

// ---- PoD slice decoding (target of review item #3) ----

func BenchmarkDecode_SliceUint64_8k(b *testing.B) {
	const l = 8192
	var buf bytes.Buffer
	e := NewBorshEncoder(&buf)
	_ = e.WriteUint32(uint32(l), LE)
	for i := 0; i < l; i++ {
		_ = e.WriteUint64(uint64(i), LE)
	}
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var got []uint64
		dec := NewBorshDecoder(data)
		_ = dec.Decode(&got)
	}
}

func BenchmarkDecode_SliceUint32_8k(b *testing.B) {
	const l = 8192
	var buf bytes.Buffer
	e := NewBorshEncoder(&buf)
	_ = e.WriteUint32(uint32(l), LE)
	for i := 0; i < l; i++ {
		_ = e.WriteUint32(uint32(i), LE)
	}
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var got []uint32
		dec := NewBorshDecoder(data)
		_ = dec.Decode(&got)
	}
}

func BenchmarkDecode_SliceUint16_8k(b *testing.B) {
	const l = 8192
	var buf bytes.Buffer
	e := NewBorshEncoder(&buf)
	_ = e.WriteUint32(uint32(l), LE)
	for i := 0; i < l; i++ {
		_ = e.WriteUint16(uint16(i), LE)
	}
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var got []uint16
		dec := NewBorshDecoder(data)
		_ = dec.Decode(&got)
	}
}

// ---- end-to-end struct encode/decode (Solana-ish layout) ----

type perfBenchStruct struct {
	A uint64
	B uint64
	C uint32
	D [32]byte
	E []uint64
}

func makePerfBenchStruct() perfBenchStruct {
	s := perfBenchStruct{A: 1, B: 2, C: 3, E: make([]uint64, 64)}
	for i := range s.E {
		s.E[i] = uint64(i)
	}
	for i := range s.D {
		s.D[i] = byte(i)
	}
	return s
}

func BenchmarkEncode_Struct_Borsh(b *testing.B) {
	s := makePerfBenchStruct()
	var w discardWriter
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := NewBorshEncoder(&w)
		_ = e.Encode(&s)
	}
}

// Buffered-mode encoder (writes into internal []byte instead of via io.Writer).
func BenchmarkEncode_Struct_Borsh_Buffered(b *testing.B) {
	s := makePerfBenchStruct()
	e := NewBorshEncoderBuf()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Reset()
		_ = e.Encode(&s)
	}
}

// MarshalBorsh goes through the pool and returns a freshly-allocated []byte
// (one alloc for the result + pool overhead). Baseline for Into.
func BenchmarkMarshal_Struct_Borsh(b *testing.B) {
	s := makePerfBenchStruct()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalBorsh(&s)
	}
}

// MarshalBorshInto writes into a caller-owned buffer: no result allocation,
// no staging buffer. Compared against BenchmarkMarshal_Struct_Borsh this
// isolates the savings from the fixed-buffer fast path.
func BenchmarkMarshalInto_Struct_Borsh(b *testing.B) {
	s := makePerfBenchStruct()
	size, err := BorshByteCount(&s)
	if err != nil {
		b.Fatal(err)
	}
	dst := make([]byte, size)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalBorshInto(&s, dst)
	}
}

// Bin-encoding variants of the two benchmarks above.
func BenchmarkMarshal_Struct_Bin(b *testing.B) {
	s := makePerfBenchStruct()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalBin(&s)
	}
}

func BenchmarkMarshalInto_Struct_Bin(b *testing.B) {
	s := makePerfBenchStruct()
	size, err := BinByteCount(&s)
	if err != nil {
		b.Fatal(err)
	}
	dst := make([]byte, size)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalBinInto(&s, dst)
	}
}

func BenchmarkEncode_WriteUint64_Buffered(b *testing.B) {
	e := NewBorshEncoderBuf()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i&0xfff == 0 {
			e.Reset()
		}
		_ = e.WriteUint64(uint64(i), LE)
	}
}

// PoD slice decode with capacity already in place — measures the cap-reuse
// fast path added in round 4.
func BenchmarkDecode_SliceUint64_8k_Reused(b *testing.B) {
	const l = 8192
	var buf bytes.Buffer
	e := NewBorshEncoder(&buf)
	_ = e.WriteUint32(uint32(l), LE)
	for i := 0; i < l; i++ {
		_ = e.WriteUint64(uint64(i), LE)
	}
	data := buf.Bytes()
	got := make([]uint64, 0, l)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewBorshDecoder(data)
		_ = dec.Decode(&got)
	}
}

// ReadString vs ReadStringBorrow — measure the unsafe.String zero-copy win.
func BenchmarkDecode_ReadString_Copy(b *testing.B) {
	payload := []byte("the quick brown fox jumps over the lazy dog")
	var buf bytes.Buffer
	e := NewBinEncoder(&buf)
	_ = e.WriteBytes(payload, true)
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewBinDecoder(data)
		_, _ = dec.ReadString()
	}
}

func BenchmarkDecode_ReadString_Borrow(b *testing.B) {
	payload := []byte("the quick brown fox jumps over the lazy dog")
	var buf bytes.Buffer
	e := NewBinEncoder(&buf)
	_ = e.WriteBytes(payload, true)
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewBinDecoder(data)
		_, _ = dec.ReadStringBorrow()
	}
}

func BenchmarkDecode_Struct_Borsh(b *testing.B) {
	s := makePerfBenchStruct()
	var buf bytes.Buffer
	_ = NewBorshEncoder(&buf).Encode(&s)
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out perfBenchStruct
		dec := NewBorshDecoder(data)
		_ = dec.Decode(&out)
	}
}
