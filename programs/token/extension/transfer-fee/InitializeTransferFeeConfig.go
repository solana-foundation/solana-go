package transferfee

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// InitializeTransferFeeConfig initializes the transfer fee on a new mint.
//
// Fails if the mint has already been initialized, so must be called before InitializeMint.
// The mint must have exactly enough space allocated for the base mint (82 bytes),
// plus 83 bytes of padding, 1 byte reserved for the account type, then space required for this extension, plus any others.
type InitializeTransferFeeConfig struct {
	// Pubkey that may update the fees (COption<Pubkey>)
	TransferFeeConfigAuthority *ag_solanago.PublicKey

	// Withdraw instructions must be signed by this key (COption<Pubkey>)
	WithdrawWithheldAuthority *ag_solanago.PublicKey

	// Amount of transfer collected as fees, expressed as basis points of the transfer amount (u16)
	TransferFeeBasisPoints *uint16

	// Maximum fee assessed on transfers (u64)
	MaximumFee *uint64

	// [0] = [WRITE] mint
	// ··········· The mint to initialize.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *InitializeTransferFeeConfig) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	if len(accounts) < 1 {
		return errors.New("not enough accounts for InitializeTransferFeeConfig")
	}
	obj.Accounts = ag_solanago.AccountMetaSlice(accounts[:1])
	return nil
}

func (slice InitializeTransferFeeConfig) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	return
}

func NewInitializeTransferFeeConfigInstructionBuilder() *InitializeTransferFeeConfig {
	nd := &InitializeTransferFeeConfig{
		Accounts: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

func (inst *InitializeTransferFeeConfig) SetTransferFeeConfigAuthority(authority ag_solanago.PublicKey) *InitializeTransferFeeConfig {
	inst.TransferFeeConfigAuthority = &authority
	return inst
}

func (inst *InitializeTransferFeeConfig) SetWithdrawWithheldAuthority(authority ag_solanago.PublicKey) *InitializeTransferFeeConfig {
	inst.WithdrawWithheldAuthority = &authority
	return inst
}

func (inst *InitializeTransferFeeConfig) SetTransferFeeBasisPoints(basisPoints uint16) *InitializeTransferFeeConfig {
	inst.TransferFeeBasisPoints = &basisPoints
	return inst
}

func (inst *InitializeTransferFeeConfig) SetMaximumFee(maxFee uint64) *InitializeTransferFeeConfig {
	inst.MaximumFee = &maxFee
	return inst
}

// SetMintAccount sets the "mint" account.
// The mint to initialize.
func (inst *InitializeTransferFeeConfig) SetMintAccount(mint ag_solanago.PublicKey) *InitializeTransferFeeConfig {
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// GetMintAccount gets the "mint" account.
// The mint to initialize.
func (inst *InitializeTransferFeeConfig) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

func (inst InitializeTransferFeeConfig) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InitializeTransferFeeConfig),
	}}
}

func (inst InitializeTransferFeeConfig) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InitializeTransferFeeConfig) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.TransferFeeConfigAuthority == nil {
			return errors.New("TransferFeeConfigAuthority parameter is not set")
		}
		if inst.WithdrawWithheldAuthority == nil {
			return errors.New("WithdrawWithheldAuthority parameter is not set")
		}
		if inst.TransferFeeBasisPoints == nil {
			return errors.New("TransferFeeBasisPoints parameter is not set")
		}
		if inst.MaximumFee == nil {
			return errors.New("MaximumFee parameter is not set")
		}
	}
	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return errors.New("accounts.Mint is not set")
		}
	}
	return nil
}

func (inst *InitializeTransferFeeConfig) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InitializeTransferFeeConfig")).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("TransferFeeConfigAuthority", inst.TransferFeeConfigAuthority))
						paramsBranch.Child(ag_format.Param("WithdrawWithheldAuthority", inst.WithdrawWithheldAuthority))
						paramsBranch.Child(ag_format.Param("TransferFeeBasisPoints", *inst.TransferFeeBasisPoints))
						paramsBranch.Child(ag_format.Param("MaximumFee", *inst.MaximumFee))
					})
					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("mint", inst.Accounts[0]))
					})
				})
		})
}

func (obj InitializeTransferFeeConfig) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `TransferFeeConfigAuthority` param:
	err = encoder.Encode(obj.TransferFeeConfigAuthority)
	if err != nil {
		return err
	}
	// Serialize `WithdrawWithheldAuthority` param:
	err = encoder.Encode(obj.WithdrawWithheldAuthority)
	if err != nil {
		return err
	}
	// Serialize `TransferFeeBasisPoints` param:
	err = encoder.Encode(obj.TransferFeeBasisPoints)
	if err != nil {
		return err
	}
	// Serialize `MaximumFee` param:
	err = encoder.Encode(obj.MaximumFee)
	if err != nil {
		return err
	}
	return nil
}

func (obj *InitializeTransferFeeConfig) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `TransferFeeConfigAuthority`:
	err = decoder.Decode(&obj.TransferFeeConfigAuthority)
	if err != nil {
		return err
	}
	// Deserialize `WithdrawWithheldAuthority`:
	err = decoder.Decode(&obj.WithdrawWithheldAuthority)
	if err != nil {
		return err
	}
	// Deserialize `TransferFeeBasisPoints`:
	err = decoder.Decode(&obj.TransferFeeBasisPoints)
	if err != nil {
		return err
	}
	// Deserialize `MaximumFee`:
	err = decoder.Decode(&obj.MaximumFee)
	if err != nil {
		return err
	}
	return nil
}
