package token2022

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// The name each confidential extension carries into errors and tree labels.
const (
	confidentialTransferName    = "ConfidentialTransfer"
	confidentialMintBurnName    = "ConfidentialMintBurn"
	confidentialTransferFeeName = "ConfidentialTransferFee"
)

// confidentialExtension is the instruction wrapper of one confidential extension.
type confidentialExtension[D confidentialSubInstructionData] interface {
	getName() string
	getSubInstructions() []confidentialSubInstruction[D]
	getSubInstruction() uint8
	getRawData() []byte
}

// Explicitly declare the extensions as implementers of the interface.
var (
	_ confidentialExtension[ConfidentialTransferSubInstructionData]    = ConfidentialTransferExtension{}
	_ confidentialExtension[ConfidentialMintBurnSubInstructionData]    = ConfidentialMintBurnExtension{}
	_ confidentialExtension[ConfidentialTransferFeeSubInstructionData] = ConfidentialTransferFeeExtension{}
)

type confidentialSubInstruction[D confidentialSubInstructionData] struct {
	name    string
	newData func() D
}

type confidentialSubInstructionData interface {
	bytes() []byte
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(b []byte) error
}

// decodeConfidentialSubInstruction outputs the typed data the extension's
// sub-instruction carries, or reports an ID the program does not define.
func decodeConfidentialSubInstruction[D confidentialSubInstructionData](
	e confidentialExtension[D],
) (D, error) {
	subInstructions := e.getSubInstructions()
	subInstruction := e.getSubInstruction()
	if int(subInstruction) >= len(subInstructions) {
		var unknown D
		return unknown, fmt.Errorf("token2022: unknown %s sub-instruction %d", e.getName(), subInstruction)
	}
	instructionData := subInstructions[subInstruction].newData()
	if err := instructionData.UnmarshalBinary(e.getRawData()); err != nil {
		var invalid D
		return invalid, err
	}
	return instructionData, nil
}

// confidentialSubInstructionLabel is the "<Extension>.<SubInstruction>" label
// the tree renderer shows, naming an ID the program does not define "Unknown".
func confidentialSubInstructionLabel[D confidentialSubInstructionData](
	e confidentialExtension[D],
) string {
	subInstructions := e.getSubInstructions()
	name := "Unknown"
	if subInstruction := e.getSubInstruction(); int(subInstruction) < len(subInstructions) {
		name = subInstructions[subInstruction].name
	}
	return e.getName() + "." + name
}

// encodeConfidentialExtensionToTree renders one confidential extension
// instruction, falling back to the payload length for data it cannot decode.
func encodeConfidentialExtensionToTree[D confidentialSubInstructionData](
	e confidentialExtension[D],
	parent treeout.Branches,
) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction(confidentialSubInstructionLabel(e))).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if data, err := decodeConfidentialSubInstruction(e); err == nil {
							paramsBranch.Child(format.Param("Data", data))
						} else {
							paramsBranch.Child(format.Param("RawData (len)", len(e.getRawData())))
						}
					})
				})
		})
}

// confidentialAuthorityAccounts appends the authority to accounts and returns
// it alongside the multisig signer metas.
func confidentialAuthorityAccounts(
	accounts solana.AccountMetaSlice,
	authority solana.PublicKey,
	multisigSigners []solana.PublicKey,
) (accountsWithAuthority, signers solana.AccountMetaSlice) {
	authorityMeta := solana.Meta(authority)
	if len(multisigSigners) == 0 {
		// The authority is the signer when multisigSigners is empty
		authorityMeta.SIGNER()
	}
	signers = make(solana.AccountMetaSlice, 0, len(multisigSigners))
	for _, signer := range multisigSigners {
		signers = append(signers, solana.Meta(signer).SIGNER())
	}
	return append(accounts, authorityMeta), signers
}

// noSubInstructionData carries the encode half of a sub-instruction that has no payload beyond the sub-instruction byte.
type noSubInstructionData struct{}

func (noSubInstructionData) bytes() []byte { return nil }

func (noSubInstructionData) MarshalBinary() ([]byte, error) { return nil, nil }

// rejectSubInstructionData reports a payload on a sub-instruction that takes none.
func rejectSubInstructionData(extension string, rawData []byte) error {
	if len(rawData) != 0 {
		return fmt.Errorf("token2022: %s sub-instruction takes no data, got %d bytes", extension, len(rawData))
	}
	return nil
}
