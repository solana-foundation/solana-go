package rpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompareCommitment(t *testing.T) {
	levels := []CommitmentType{CommitmentProcessed, CommitmentConfirmed, CommitmentFinalized}
	for i, a := range levels {
		for j, b := range levels {
			got := CompareCommitment(a, b)
			switch {
			case i < j:
				require.Negative(t, got, "%s < %s", a, b)
			case i > j:
				require.Positive(t, got, "%s > %s", a, b)
			default:
				require.Zero(t, got, "%s == %s", a, b)
			}
		}
	}

	// A status without a known level must never satisfy a required one.
	for _, unknown := range []CommitmentType{"", "recent", "max"} {
		for _, level := range levels {
			require.Negative(t, CompareCommitment(unknown, level), "%q < %s", unknown, level)
			require.Positive(t, CompareCommitment(level, unknown), "%s > %q", level, unknown)
		}
	}
	require.Zero(t, CompareCommitment("", ""))
}
