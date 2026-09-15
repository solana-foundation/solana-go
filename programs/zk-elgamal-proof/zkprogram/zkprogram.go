// Package zkprogram provides the API for constructing clients for the ZK ElGamal program.
package zkprogram

import (
	"github.com/solana-foundation/solana-go/v2"
)

const ProgramName = "ZkElGamalProof"

var ProgramID solana.PublicKey = solana.ZKElGamalProofProgramID

// SetProgramID points the package at the program deployed under a different
// id.
func SetProgramID(pubkey solana.PublicKey) error {
	ProgramID = pubkey
	return solana.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

func init() {
	if !ProgramID.IsZero() {
		solana.MustRegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
	}
}
