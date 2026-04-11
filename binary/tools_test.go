package bin

import "encoding/binary"

func concatByteSlices(slices ...[]byte) (out []byte) {
	for i := range slices {
		out = append(out, slices[i]...)
	}
	return
}
func uint16ToBytes(i uint16, order binary.ByteOrder) []byte {
	buf := make([]byte, 2)
	order.PutUint16(buf, i)
	return buf
}

func uint32ToBytes(i uint32, order binary.ByteOrder) []byte {
	buf := make([]byte, 4)
	order.PutUint32(buf, i)
	return buf
}

func uint64ToBytes(i uint64, order binary.ByteOrder) []byte {
	buf := make([]byte, 8)
	order.PutUint64(buf, i)
	return buf
}
