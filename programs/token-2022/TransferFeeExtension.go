package token2022

import (
	"encoding/binary"
	"errors"
	"fmt"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// TransferFee sub-instruction IDs.
const (
	TransferFee_InitializeTransferFeeConfig uint8 = iota
	TransferFee_TransferCheckedWithFee
	TransferFee_WithdrawWithheldTokensFromMint
	TransferFee_WithdrawWithheldTokensFromAccounts
	TransferFee_HarvestWithheldTokensToMint
	TransferFee_SetTransferFee
)

// TransferFeeExtension is the instruction wrapper for the TransferFee extension (ID 26).
type TransferFeeExtension struct {
	SubInstruction uint8

	// Sub-instruction data (only the relevant fields are set based on SubInstruction).
	TransferFeeConfigAuthority *ag_solanago.PublicKey `bin:"-"`
	WithdrawWithheldAuthority  *ag_solanago.PublicKey `bin:"-"`
	TransferFeeBasisPoints     uint16                 `bin:"-"`
	MaximumFee                 uint64                 `bin:"-"`

	// For TransferCheckedWithFee:
	Amount   *uint64 `bin:"-"`
	Decimals *uint8  `bin:"-"`
	Fee      *uint64 `bin:"-"`

	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *TransferFeeExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	switch obj.SubInstruction {
	case TransferFee_InitializeTransferFeeConfig:
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	case TransferFee_TransferCheckedWithFee:
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(4)
	case TransferFee_WithdrawWithheldTokensFromMint:
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	case TransferFee_WithdrawWithheldTokensFromAccounts:
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(3)
	case TransferFee_HarvestWithheldTokensToMint:
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	case TransferFee_SetTransferFee:
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	default:
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	}
	return nil
}

func (slice TransferFeeExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst TransferFeeExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_TransferFeeExtension),
	}}
}

func (inst TransferFeeExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *TransferFeeExtension) Validate() error {
	if len(inst.Accounts) == 0 || inst.Accounts[0] == nil {
		return errors.New("accounts[0] is not set")
	}
	switch inst.SubInstruction {
	case TransferFee_TransferCheckedWithFee:
		if inst.Amount == nil {
			return errors.New("amount parameter is not set")
		}
		if inst.Decimals == nil {
			return errors.New("decimals parameter is not set")
		}
		if inst.Fee == nil {
			return errors.New("fee parameter is not set")
		}
		for i := 0; i < 4; i++ {
			if len(inst.Accounts) <= i || inst.Accounts[i] == nil {
				return fmt.Errorf("accounts[%d] is not set", i)
			}
		}
		if !inst.Accounts[3].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	case TransferFee_WithdrawWithheldTokensFromMint:
		for i := 0; i < 3; i++ {
			if len(inst.Accounts) <= i || inst.Accounts[i] == nil {
				return fmt.Errorf("accounts[%d] is not set", i)
			}
		}
		if !inst.Accounts[2].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	case TransferFee_WithdrawWithheldTokensFromAccounts:
		for i := 0; i < 3; i++ {
			if len(inst.Accounts) <= i || inst.Accounts[i] == nil {
				return fmt.Errorf("accounts[%d] is not set", i)
			}
		}
		if !inst.Accounts[2].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	case TransferFee_SetTransferFee:
		for i := 0; i < 2; i++ {
			if len(inst.Accounts) <= i || inst.Accounts[i] == nil {
				return fmt.Errorf("accounts[%d] is not set", i)
			}
		}
		if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	}
	if len(inst.Signers) > MAX_SIGNERS {
		return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
	}
	return nil
}

func (inst *TransferFeeExtension) EncodeToTree(parent ag_treeout.Branches) {
	names := []string{
		"InitializeTransferFeeConfig",
		"TransferCheckedWithFee",
		"WithdrawWithheldTokensFromMint",
		"WithdrawWithheldTokensFromAccounts",
		"HarvestWithheldTokensToMint",
		"SetTransferFee",
	}
	name := "Unknown"
	if int(inst.SubInstruction) < len(names) {
		name = names[inst.SubInstruction]
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("TransferFee." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						switch inst.SubInstruction {
						case TransferFee_InitializeTransferFeeConfig:
							paramsBranch.Child(ag_format.Param("TransferFeeBasisPoints", inst.TransferFeeBasisPoints))
							paramsBranch.Child(ag_format.Param("           MaximumFee", inst.MaximumFee))
						case TransferFee_TransferCheckedWithFee:
							paramsBranch.Child(ag_format.Param("  Amount", *inst.Amount))
							paramsBranch.Child(ag_format.Param("Decimals", *inst.Decimals))
							paramsBranch.Child(ag_format.Param("     Fee", *inst.Fee))
						case TransferFee_SetTransferFee:
							paramsBranch.Child(ag_format.Param("TransferFeeBasisPoints", inst.TransferFeeBasisPoints))
							paramsBranch.Child(ag_format.Param("           MaximumFee", inst.MaximumFee))
						}
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						for i, acct := range inst.Accounts {
							if acct != nil {
								accountsBranch.Child(ag_format.Meta(fmt.Sprintf("[%d]", i), acct))
							}
						}
					})
				})
		})
}

func (obj TransferFeeExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.WriteUint8(obj.SubInstruction)
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case TransferFee_InitializeTransferFeeConfig:
		if err = writeOptionalPubkey(encoder, obj.TransferFeeConfigAuthority); err != nil {
			return err
		}
		if err = writeOptionalPubkey(encoder, obj.WithdrawWithheldAuthority); err != nil {
			return err
		}
		err = encoder.WriteUint16(obj.TransferFeeBasisPoints, binary.LittleEndian)
		if err != nil {
			return err
		}
		err = encoder.WriteUint64(obj.MaximumFee, binary.LittleEndian)
		if err != nil {
			return err
		}
	case TransferFee_TransferCheckedWithFee:
		err = encoder.WriteUint64(*obj.Amount, binary.LittleEndian)
		if err != nil {
			return err
		}
		err = encoder.WriteUint8(*obj.Decimals)
		if err != nil {
			return err
		}
		err = encoder.WriteUint64(*obj.Fee, binary.LittleEndian)
		if err != nil {
			return err
		}
	case TransferFee_SetTransferFee:
		err = encoder.WriteUint16(obj.TransferFeeBasisPoints, binary.LittleEndian)
		if err != nil {
			return err
		}
		err = encoder.WriteUint64(obj.MaximumFee, binary.LittleEndian)
		if err != nil {
			return err
		}
		// WithdrawWithheldTokensFromMint, WithdrawWithheldTokensFromAccounts,
		// HarvestWithheldTokensToMint have no additional data
	}
	return nil
}

func (obj *TransferFeeExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case TransferFee_InitializeTransferFeeConfig:
		obj.TransferFeeConfigAuthority, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
		obj.WithdrawWithheldAuthority, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
		obj.TransferFeeBasisPoints, err = decoder.ReadUint16(binary.LittleEndian)
		if err != nil {
			return err
		}
		obj.MaximumFee, err = decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
	case TransferFee_TransferCheckedWithFee:
		amount, err := decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		obj.Amount = &amount
		decimals, err := decoder.ReadUint8()
		if err != nil {
			return err
		}
		obj.Decimals = &decimals
		fee, err := decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		obj.Fee = &fee
	case TransferFee_SetTransferFee:
		obj.TransferFeeBasisPoints, err = decoder.ReadUint16(binary.LittleEndian)
		if err != nil {
			return err
		}
		obj.MaximumFee, err = decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewInitializeTransferFeeConfigInstruction creates an instruction to initialize the transfer fee configuration.
func NewInitializeTransferFeeConfigInstruction(
	transferFeeConfigAuthority *ag_solanago.PublicKey,
	withdrawWithheldAuthority *ag_solanago.PublicKey,
	transferFeeBasisPoints uint16,
	maximumFee uint64,
	mint ag_solanago.PublicKey,
) *TransferFeeExtension {
	inst := &TransferFeeExtension{
		SubInstruction:             TransferFee_InitializeTransferFeeConfig,
		TransferFeeConfigAuthority: transferFeeConfigAuthority,
		WithdrawWithheldAuthority:  withdrawWithheldAuthority,
		TransferFeeBasisPoints:     transferFeeBasisPoints,
		MaximumFee:                 maximumFee,
		Accounts:                   make(ag_solanago.AccountMetaSlice, 1),
		Signers:                    make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// NewTransferCheckedWithFeeInstruction creates an instruction to transfer tokens with fee.
func NewTransferCheckedWithFeeInstruction(
	amount uint64,
	decimals uint8,
	fee uint64,
	source ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	destination ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *TransferFeeExtension {
	inst := &TransferFeeExtension{
		SubInstruction: TransferFee_TransferCheckedWithFee,
		Amount:         &amount,
		Decimals:       &decimals,
		Fee:            &fee,
		Accounts:       make(ag_solanago.AccountMetaSlice, 4),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(source).WRITE()
	inst.Accounts[1] = ag_solanago.Meta(mint)
	inst.Accounts[2] = ag_solanago.Meta(destination).WRITE()
	inst.Accounts[3] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[3].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// NewWithdrawWithheldTokensFromMintInstruction creates an instruction to withdraw withheld tokens from the mint.
func NewWithdrawWithheldTokensFromMintInstruction(
	mint ag_solanago.PublicKey,
	destination ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *TransferFeeExtension {
	inst := &TransferFeeExtension{
		SubInstruction: TransferFee_WithdrawWithheldTokensFromMint,
		Accounts:       make(ag_solanago.AccountMetaSlice, 3),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	inst.Accounts[2] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// NewWithdrawWithheldTokensFromAccountsInstruction creates an instruction to withdraw withheld tokens
// from specific token accounts.
func NewWithdrawWithheldTokensFromAccountsInstruction(
	mint ag_solanago.PublicKey,
	destination ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
	sourceAccounts ...ag_solanago.PublicKey,
) *TransferFeeExtension {
	inst := &TransferFeeExtension{
		SubInstruction: TransferFee_WithdrawWithheldTokensFromAccounts,
		Accounts:       make(ag_solanago.AccountMetaSlice, 3+len(sourceAccounts)),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint)
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	inst.Accounts[2] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[2].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	for i, src := range sourceAccounts {
		inst.Accounts[3+i] = ag_solanago.Meta(src).WRITE()
	}
	return inst
}

// NewHarvestWithheldTokensToMintInstruction creates an instruction to harvest withheld tokens to the mint.
func NewHarvestWithheldTokensToMintInstruction(
	mint ag_solanago.PublicKey,
	sourceAccounts ...ag_solanago.PublicKey,
) *TransferFeeExtension {
	inst := &TransferFeeExtension{
		SubInstruction: TransferFee_HarvestWithheldTokensToMint,
		Accounts:       make(ag_solanago.AccountMetaSlice, 1+len(sourceAccounts)),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	for i, src := range sourceAccounts {
		inst.Accounts[1+i] = ag_solanago.Meta(src).WRITE()
	}
	return inst
}

// NewSetTransferFeeInstruction creates an instruction to set the transfer fee.
func NewSetTransferFeeInstruction(
	transferFeeBasisPoints uint16,
	maximumFee uint64,
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *TransferFeeExtension {
	inst := &TransferFeeExtension{
		SubInstruction:         TransferFee_SetTransferFee,
		TransferFeeBasisPoints: transferFeeBasisPoints,
		MaximumFee:             maximumFee,
		Accounts:               make(ag_solanago.AccountMetaSlice, 2),
		Signers:                make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	inst.Accounts[1] = ag_solanago.Meta(authority)
	if len(multisigSigners) == 0 {
		inst.Accounts[1].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}
