package solend

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Flash repays liquidity from a reserve and transfers it to a destination liquidity account.
type FlashRepayReserveLiquidity struct {
	// The amount of tokens to flash repay.
	Amount                 *uint64
	BorrowInstructionIndex *uint8

	// [0] = [WRITE] source_liquidity
	// ··········· $authority can transfer $liquidity_amount.
	//
	// [1] = [WRITE] destination_liquidity
	//
	// [2] = [WRITE] flash_loan_fee_receiver
	// ··········· Must match the reserve liquidity fee receiver.
	//
	// [3] = [WRITE] host_fee_receiver
	//
	// [4] = [WRITE] reserve
	//
	// [5] = [] lending_market
	//
	// [6] = [SIGNER] user_transfer_authority ($authority)
	//
	// [7] = [] sysvar_instructions
	//
	// [8] = [] token_program
	// ··········· The SPL Token program account (read-only).
	//
	// [9...] = [SIGNER] signers
	// ··········· M signer accounts.
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *FlashRepayReserveLiquidity) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(9)
	return nil
}

func (slice FlashRepayReserveLiquidity) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

// NewFlashRepayReserveLiquidityInstructionBuilder creates a new `FlashRepayReserveLiquidity` instruction builder.
func NewFlashRepayReserveLiquidityInstructionBuilder() *FlashRepayReserveLiquidity {
	nd := &FlashRepayReserveLiquidity{
		Accounts: make(ag_solanago.AccountMetaSlice, 9),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetAmount sets the "amount" parameter.
// The amount of tokens to transfer.
func (inst *FlashRepayReserveLiquidity) SetAmount(amount uint64) *FlashRepayReserveLiquidity {
	inst.Amount = &amount
	return inst
}

// SetBorrowInstructionIndex sets the "borrow_instruction_index" parameter.
// The index of the borrow instruction in the borrow instruction list.
func (inst *FlashRepayReserveLiquidity) SetBorrowInstructionIndex(borrowInstructionIndex uint8) *FlashRepayReserveLiquidity {
	inst.BorrowInstructionIndex = &borrowInstructionIndex
	return inst
}

// SetSourceAccount sets the "source" account.
// The source account.
func (inst *FlashRepayReserveLiquidity) SetSourceAccount(source ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[0] = ag_solanago.Meta(source).WRITE()
	return inst
}

// GetSourceAccount gets the "source" account.
// The source account.
func (inst *FlashRepayReserveLiquidity) GetSourceAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetDestinationAccount sets the "destination" account.
// The destination account.
func (inst *FlashRepayReserveLiquidity) SetDestinationAccount(destination ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	return inst
}

// GetDestinationAccount gets the "destination" account.
// The destination account.
func (inst *FlashRepayReserveLiquidity) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

// SetFlashLoanFeeReceiver sets the "flash_loan_fee_receiver" account.
// The flash loan fee receiver account.
func (inst *FlashRepayReserveLiquidity) SetFlashLoanFeeReceiver(flashLoanFeeReceiver ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[2] = ag_solanago.Meta(flashLoanFeeReceiver).WRITE()
	return inst
}

// GetFlashLoanFeeReceiver gets the "flash_loan_fee_receiver" account.
// The flash loan fee receiver account.
func (inst *FlashRepayReserveLiquidity) GetFlashLoanFeeReceiver() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

// SetHostFeeReceiver sets the "host_fee_receiver" account.
// The host fee receiver account.
func (inst *FlashRepayReserveLiquidity) SetHostFeeReceiver(hostFeeReceiver ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[3] = ag_solanago.Meta(hostFeeReceiver).WRITE()
	return inst
}

// GetHostFeeReceiver gets the "host_fee_receiver" account.
// The host fee receiver account.
func (inst *FlashRepayReserveLiquidity) GetHostFeeReceiver() *ag_solanago.AccountMeta {
	return inst.Accounts[3]
}

// SetReserveAccount sets the "reserve" account.
// The reserve account.
func (inst *FlashRepayReserveLiquidity) SetReserveAccount(reserve ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[4] = ag_solanago.Meta(reserve).WRITE()
	return inst
}

// GetReserveAccount gets the "reserve" account.
// The reserve account.
func (inst *FlashRepayReserveLiquidity) GetReserveAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[4]
}

// SetLendingMarketAccount sets the "lending_market" account.
// The lending market account (read-only).
func (inst *FlashRepayReserveLiquidity) SetLendingMarketAccount(lendingMarket ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[5] = ag_solanago.Meta(lendingMarket)
	return inst
}

// GetLendingMarketAccount gets the "lending_market" account.
// The lending market account (read-only).
func (inst *FlashRepayReserveLiquidity) GetLendingMarketAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[5]
}

// SetLendingMarketAuthorityAccount sets the "lending_market_authority" account.
// The lending market authority account (read-only).
func (inst *FlashRepayReserveLiquidity) SetTransferAuthorityAccount(transferAuthority ag_solanago.PublicKey, multisigSigners ...ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[6] = ag_solanago.Meta(transferAuthority)
	if len(multisigSigners) == 0 {
		inst.Accounts[6].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}

// GetTransferAuthorityAccount gets the "transfer_authority" account.
// The transfer authority account.
func (inst *FlashRepayReserveLiquidity) GetTransferAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[6]
}

// SetSysvarInstructionsAccount sets the "sysvar_instructions" account.
// The sysvar instructions account (read-only).
func (inst *FlashRepayReserveLiquidity) SetSysvarInstructionsAccount(sysvarInstructions ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[7] = ag_solanago.Meta(sysvarInstructions)
	return inst
}

// GetSysvarInstructionsAccount gets the "sysvar_instructions" account.
// The sysvar instructions account (read-only).
func (inst *FlashRepayReserveLiquidity) GetSysvarInstructionsAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[7]
}

// SetTokenProgramAccount sets the "token_program" account.
// The SPL Token program account (read-only).
func (inst *FlashRepayReserveLiquidity) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *FlashRepayReserveLiquidity {
	inst.Accounts[8] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "token_program" account.
// The SPL Token program account (read-only).
func (inst *FlashRepayReserveLiquidity) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[8]
}

func (inst FlashRepayReserveLiquidity) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_FlashRepayReserveLiquidity),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst FlashRepayReserveLiquidity) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *FlashRepayReserveLiquidity) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Amount == nil {
			return errors.New("Amount parameter is not set")
		}
		if inst.BorrowInstructionIndex == nil {
			return errors.New("BorrowInstructionIndex parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.Accounts[0] == nil {
			return fmt.Errorf("accounts.SourceLiquidity is not set")
		}
		if inst.Accounts[1] == nil {
			return fmt.Errorf("accounts.DestinationLiquidity is not set")
		}
		if inst.Accounts[2] == nil {
			return fmt.Errorf("accounts.FlashLoanFeeReceiver is not set")
		}
		if inst.Accounts[3] == nil {
			return fmt.Errorf("accounts.HostFeeReceiver is not set")
		}
		if inst.Accounts[4] == nil {
			return fmt.Errorf("accounts.Reserve is not set")
		}
		if inst.Accounts[5] == nil {
			return fmt.Errorf("accounts.LendingMarket is not set")
		}
		if inst.Accounts[6] == nil {
			return fmt.Errorf("accounts.TransferAuthority is not set")
		}
		if !inst.Accounts[6].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
		if inst.Accounts[7] == nil {
			return fmt.Errorf("accounts.SysvarInstructions is not set")
		}
		if inst.Accounts[8] == nil {
			return fmt.Errorf("accounts.TokenProgram is not set")
		}
		if len(inst.Signers) > MAX_SIGNERS {
			return fmt.Errorf("too many signers; got %v, but max is 11", len(inst.Signers))
		}
	}
	return nil
}

func (inst *FlashRepayReserveLiquidity) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("FlashRepayReserveLiquidity")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Amount", *inst.Amount))
						paramsBranch.Child(ag_format.Param("BorrowInstructionIndex", *inst.BorrowInstructionIndex))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("source_liquidity", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("destination_liquidity", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("flash_loan_fee_receiver", inst.Accounts[2]))
						accountsBranch.Child(ag_format.Meta("host_fee_receiver", inst.Accounts[3]))
						accountsBranch.Child(ag_format.Meta("reserve", inst.Accounts[4]))
						accountsBranch.Child(ag_format.Meta("lending_market", inst.Accounts[5]))
						accountsBranch.Child(ag_format.Meta("lending_market_authority", inst.Accounts[6]))
						accountsBranch.Child(ag_format.Meta("sysvar_instructions", inst.Accounts[7]))
						accountsBranch.Child(ag_format.Meta("token_program", inst.Accounts[8]))

						signersBranch := accountsBranch.Child(fmt.Sprintf("signers[len=%v]", len(inst.Signers)))
						for i, v := range inst.Signers {
							if len(inst.Signers) > 9 && i < 10 {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf(" [%v]", i), v))
							} else {
								signersBranch.Child(ag_format.Meta(fmt.Sprintf("[%v]", i), v))
							}
						}
					})
				})
		})
}

func (inst FlashRepayReserveLiquidity) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	// Serialize `Amount` param:
	if err := encoder.Encode(inst.Amount); err != nil {
		return err
	}
	// Serialize `BorrowInstructionIndex` param:
	if err := encoder.Encode(inst.BorrowInstructionIndex); err != nil {
		return err
	}
	return nil
}

func (inst *FlashRepayReserveLiquidity) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	// Deserialize `Amount`:
	if err := decoder.Decode(&inst.Amount); err != nil {
		return err
	}
	// Deserialize `BorrowInstructionIndex`:
	if err := decoder.Decode(&inst.BorrowInstructionIndex); err != nil {
		return err
	}
	return nil
}

func NewFlashRepayReserveLiquidityInstruction(
	amount uint64,
	borrowInstructionIndex uint8,
	sourceLiquidity ag_solanago.PublicKey,
	destinationLiquidity ag_solanago.PublicKey,
	flashLoanFeeReceiver ag_solanago.PublicKey,
	hostFeeReceiver ag_solanago.PublicKey,
	reserve ag_solanago.PublicKey,
	lendingMarket ag_solanago.PublicKey,
	transferAuthority ag_solanago.PublicKey,
	sysvarInstructions ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *FlashRepayReserveLiquidity {
	return NewFlashRepayReserveLiquidityInstructionBuilder().
		SetAmount(amount).
		SetBorrowInstructionIndex(borrowInstructionIndex).
		SetSourceAccount(sourceLiquidity).
		SetDestinationAccount(destinationLiquidity).
		SetFlashLoanFeeReceiver(flashLoanFeeReceiver).
		SetHostFeeReceiver(hostFeeReceiver).
		SetReserveAccount(reserve).
		SetLendingMarketAccount(lendingMarket).
		SetTransferAuthorityAccount(transferAuthority, multisigSigners...).
		SetSysvarInstructionsAccount(sysvarInstructions).
		SetTokenProgramAccount(tokenProgram)
}
