package token2022

import (
	"fmt"
	"math"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/encryption"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/proofdata"
	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/zkprogram"
)

// NewConfidentialTransferFeeWithdrawWithheldTokensFromAccountsInstructions
// transfers the withheld confidential tokens of the source accounts to an
// account, appending the verification instruction when the proof is in a
// sibling instruction.
//
// The proof must sit directly after the withdraw instruction; for any other
// offset use NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromAccountsInstruction.
//
// This instruction is susceptible to front-running: the proof is verified
// against the fees withheld at the source accounts, so a fee arriving before
// the transaction lands invalidates it. HarvestWithheldTokensToMint followed by
// WithdrawWithheldTokensFromMint is the alternative.
func NewConfidentialTransferFeeWithdrawWithheldTokensFromAccountsInstructions(
	mint solana.PublicKey,
	destination solana.PublicKey,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	sources []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData],
) ([]solana.Instruction, error) {
	inner, err := NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromAccountsInstruction(
		mint, destination, newDecryptableAvailableBalance, authority, multisigSigners, sources, proofDataLocation)
	if err != nil {
		return nil, err
	}
	built, err := inner.ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	return appendVerifyProofInstruction([]solana.Instruction{built},
		zkprogram.VerifyCiphertextCiphertextEquality, proofDataLocation)
}

// NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromAccountsInstruction
// builds the WithdrawWithheldTokensFromAccounts instruction on its own, for use
// with a cross-program invoke.
func NewConfidentialTransferFeeInnerWithdrawWithheldTokensFromAccountsInstruction(
	mint solana.PublicKey,
	destination solana.PublicKey,
	newDecryptableAvailableBalance encryption.AeCiphertext,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
	sources []solana.PublicKey,
	proofDataLocation zkprogram.ProofLocation[*proofdata.CiphertextCiphertextEqualityProofData],
) (*ConfidentialTransferFeeExtension, error) {
	if len(sources) > math.MaxUint8 {
		return nil, fmt.Errorf("token2022: WithdrawWithheldTokensFromAccounts takes at most %d source accounts, got %d", math.MaxUint8, len(sources))
	}
	proofAccount, proofInstructionOffset, err := resolveProofLocation(proofDataLocation)
	if err != nil {
		return nil, err
	}
	data := ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData{
		NumTokenAccounts:               uint8(len(sources)),
		ProofInstructionOffset:         proofInstructionOffset,
		NewDecryptableAvailableBalance: newDecryptableAvailableBalance,
	}
	accounts, signers := confidentialAuthorityAccounts(
		solana.AccountMetaSlice{
			solana.Meta(mint).WRITE(),
			solana.Meta(destination).WRITE(),
			proofAccount,
		},
		authority,
		multisigSigners,
	)
	// The source accounts must trail the multisig signers,
	// so the signers are inlined into Accounts and Signers stays empty (see GetAccounts).
	accounts = append(accounts, signers...)
	for _, source := range sources {
		accounts = append(accounts, solana.Meta(source).WRITE())
	}
	return &ConfidentialTransferFeeExtension{
		SubInstruction: ConfidentialTransferFee_WithdrawWithheldTokensFromAccounts,
		RawData:        data.bytes(),
		Accounts:       accounts,
		Signers:        make(solana.AccountMetaSlice, 0),
	}, nil
}

// ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData is the
// instruction data for ConfidentialTransferFee_WithdrawWithheldTokensFromAccounts.
type ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData struct {
	// NumTokenAccounts is the number of source accounts to withdraw from.
	NumTokenAccounts uint8
	// ProofInstructionOffset locates the VerifyCiphertextCiphertextEquality
	// instruction relative to the WithdrawWithheldTokensFromAccounts
	// instruction; zero means a context state account.
	ProofInstructionOffset int8
	// NewDecryptableAvailableBalance is the new decryptable balance in the
	// destination token account.
	NewDecryptableAvailableBalance encryption.AeCiphertext
}

const confidentialTransferFeeWithdrawWithheldTokensFromAccountsDataSize = u8Size + proofOffsetSize + aeCiphertextSize

func (d ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData) bytes() []byte {
	out := make([]byte, 0, confidentialTransferFeeWithdrawWithheldTokensFromAccountsDataSize)
	out = append(out, d.NumTokenAccounts)
	out = append(out, byte(d.ProofInstructionOffset))
	out = append(out, d.NewDecryptableAvailableBalance[:]...)
	return out
}

func (d ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData) MarshalBinary() ([]byte, error) {
	return d.bytes(), nil
}

func (d *ConfidentialTransferFeeWithdrawWithheldTokensFromAccountsData) UnmarshalBinary(b []byte) error {
	if len(b) != confidentialTransferFeeWithdrawWithheldTokensFromAccountsDataSize {
		return fmt.Errorf("token2022: ConfidentialTransferFee WithdrawWithheldTokensFromAccounts data is %d bytes, want %d", len(b), confidentialTransferFeeWithdrawWithheldTokensFromAccountsDataSize)
	}
	d.NumTokenAccounts = b[0]
	d.ProofInstructionOffset = int8(b[1])
	copy(d.NewDecryptableAvailableBalance[:], b[2:])
	return nil
}
