package token2022

import (
	"bytes"

	"github.com/gagliardetto/solana-go"
)

type programInstruction byte

const (
	initialize programInstruction = 0
	update     programInstruction = 1
)

var (
	/*
		The Token-2022 Program, also known as Token Extensions, is a superset of the functionality provided by the Token Program.

		For more information, see the [Token-2022 Program Documentation](https://spl.solana.com/token-2022).
	*/
	ProgramID = solana.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
)

type TokenInstruction byte

const (
	InitializeMint TokenInstruction = iota
	InitializeAccount
	InitializeMultisig
	Transfer
	Approve
	Revoke
	SetAuthority
	MintTo
	Burn
	CloseAccount
	FreezeAccount
	ThawAccount
	TransferChecked
	ApproveChecked
	MintToChecked
	BurnChecked
	InitializeAccount2
	SyncNative
	InitializeAccount3
	InitializeMultisig2
	InitializeMint2
	GetAccountDataSize
	InitializeImmutableOwner
	AmountToUiAmount
	UiAmountToAmount
	InitializeMintCloseAuthority
	TransferFeeExtension
	ConfidentialTransferExtension
	DefaultAccountStateExtension
	Reallocate
	MemoTransferExtension
	CreateNativeMint
	InitializeNonTransferableMint
	InterestBearingMintExtension
	CpiGuardExtension
	InitializePermanentDelegate
	TransferHookExtension
	ConfidentialTransferFeeExtension
	WithdrawExcessLamports
	MetadataPointerExtension
	GroupPointerExtension
	GroupMemberPointerExtension
	ConfidentialMintBurnExtension
	ScaledUiAmountExtension
	PausableExtension
)

type instruction struct {
	programID solana.PublicKey
	accounts  []*solana.AccountMeta
	data      []byte
}

func (inst *instruction) ProgramID() solana.PublicKey {
	return inst.programID
}

func (inst *instruction) Accounts() (out []*solana.AccountMeta) {
	return inst.accounts
}

func (inst *instruction) Data() ([]byte, error) {
	return inst.data, nil
}

type InitializeInstructionArgs struct {
	Metadata        solana.PublicKey
	UpdateAuthority solana.PublicKey
	Mint            solana.PublicKey
	MintAuthority   solana.PublicKey
	Name            string
	Symbol          string
	Uri             string
}

func CreateInitializeInstruction(
	args InitializeInstructionArgs,
) solana.Instruction {
	programID := ProgramID

	ix := &instruction{
		programID: programID,
		accounts: []*solana.AccountMeta{
			solana.Meta(args.Metadata).WRITE(),
			solana.Meta(args.UpdateAuthority),
			solana.Meta(args.Mint),
			solana.Meta(args.MintAuthority).SIGNER(),
		},
		data: encodeInitializeInstructionData(
			args.Name,
			args.Symbol,
			args.Uri,
		),
	}

	return ix
}

func encodeInitializeInstructionData(
	name string,
	symbol string,
	uri string,
) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{210, 225, 30, 162, 88, 184, 77, 141, byte(len([]byte(name))), 0, 0, 0})
	buf.Write([]byte(name))
	buf.Write([]byte{byte(len([]byte(symbol))), 0, 0, 0})
	buf.Write([]byte(symbol))
	buf.Write([]byte{byte(len([]byte(uri))), 0, 0, 0})
	buf.Write([]byte(uri))

	return buf.Bytes()
}
