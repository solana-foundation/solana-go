package vote

import (
	"fmt"

	bin "github.com/gagliardetto/binary"
)

// checkedCount rejects an element count that cannot possibly fit in the bytes
// still to be read, before it is used to size an allocation. Counts come from
// the wire (instruction data, account data) and are attacker-controlled:
// without this bound a 12-byte payload can request a multi-GiB slice or
// panic with "makeslice: len out of range".
func checkedCount(dec *bin.Decoder, count uint64, minElemSize int, what string) (int, error) {
	if count > uint64(dec.Remaining())/uint64(minElemSize) {
		return 0, fmt.Errorf("%s count %d exceeds remaining %d bytes (min %d bytes each)", what, count, dec.Remaining(), minElemSize)
	}
	return int(count), nil
}
