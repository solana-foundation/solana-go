package stake

import (
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
)

func pubkeyOf(v byte) solana.PublicKey {
	var pk solana.PublicKey
	for i := range pk {
		pk[i] = v
	}
	return pk
}

func u32LE(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func u64LE(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

func concat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func encodeInst(inst interface{ Build() *Instruction }) ([]byte, error) {
	built := inst.Build()
	return built.Data()
}
