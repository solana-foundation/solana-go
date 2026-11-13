package vote

import (
	"testing"

	bin "github.com/gagliardetto/binary"
)

// A V1_14_11 vote-account image up to the Votes count:
// version u32=1 | NodePubkey[32] | AuthorizedWithdrawer[32] | Commission u8 | count u64.
func voteAccountV1_14_11(count uint64) []byte {
	b := []byte{0x01, 0x00, 0x00, 0x00}
	b = append(b, make([]byte, 32)...)
	b = append(b, make([]byte, 32)...)
	b = append(b, 0x00)
	c := make([]byte, 8)
	for i := 0; i < 8; i++ {
		c[i] = byte(count >> (8 * uint(i)))
	}
	return append(b, c...)
}

// DecodeVoteAccount must reject a vote count larger than the remaining bytes
// instead of panicking in make([]Lockout, count). Before the bound this panics
// with "makeslice: len out of range".
func TestDecodeVoteAccountRejectsHugeVoteCount(t *testing.T) {
	if _, err := DecodeVoteAccount(voteAccountV1_14_11(^uint64(0))); err == nil {
		t.Fatal("expected an error for a vote count beyond the remaining data, got nil")
	}
}

// Same for the exported AuthorizedVoters decoder: an 8-byte count must not drive
// make([]AuthorizedVoter, count).
func TestAuthorizedVotersRejectsHugeCount(t *testing.T) {
	var av AuthorizedVoters
	err := av.UnmarshalWithDecoder(bin.NewBinDecoder([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}))
	if err == nil {
		t.Fatal("expected an error for an authorized-voter count beyond the remaining data, got nil")
	}
}

// The bound must not reject a well-formed vector: count=1 followed by one 40-byte voter.
func TestAuthorizedVotersAcceptsValidCount(t *testing.T) {
	buf := make([]byte, 8+40)
	buf[0] = 1
	var av AuthorizedVoters
	if err := av.UnmarshalWithDecoder(bin.NewBinDecoder(buf)); err != nil {
		t.Fatalf("valid authorized voters failed to decode: %v", err)
	}
	if len(av.Voters) != 1 {
		t.Fatalf("expected 1 voter, got %d", len(av.Voters))
	}
}
