package solend

import (
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Flash borrows liquidity from a reserve and transfers it to a destination liquidity account.
type FlashBorrowReserveLiquidity struct {
	// The amount of tokens to flash borrow.
	Amount *uint64

	// [0] = [WRITE] source_liquidity
	// ··········· The source liquidity account.
	//
	// [1] = [WRITE] destination_liquidity
	// ··········· The destination liquidity account.
	//
	// [2] = [WRITE] reserve
	// ··········· The reserve account.
	//
	// [3] = [] lending_market
	// ··········· The lending market account (read-only).
	//
	// [4] = [] lending_market_authority
	// ··········· The lending market authority account (read-only).
	//
	// [5] = [] sysvar_instructions
	// ··········· The sysvar instructions account (read-only).
	//
	// [6] = [] token_program
	// ··········· The SPL Token program account (read-only).
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *FlashBorrowReserveLiquidity) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(7)
	return nil
}

func (slice FlashBorrowReserveLiquidity) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

// NewFlashBorrowReserveLiquidityInstructionBuilder creates a new `FlashBorrowReserveLiquidity` instruction builder.
func NewFlashBorrowReserveLiquidityInstructionBuilder() *FlashBorrowReserveLiquidity {
	nd := &FlashBorrowReserveLiquidity{
		Accounts: make(ag_solanago.AccountMetaSlice, 7),
		Signers:  make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

// SetAmount sets the "amount" parameter.
// The amount of tokens to transfer.
func (inst *FlashBorrowReserveLiquidity) SetAmount(amount uint64) *FlashBorrowReserveLiquidity {
	inst.Amount = &amount
	return inst
}

// SetSourceAccount sets the "source" account.
// The source account.
func (inst *FlashBorrowReserveLiquidity) SetSourceAccount(source ag_solanago.PublicKey) *FlashBorrowReserveLiquidity {
	inst.Accounts[0] = ag_solanago.Meta(source).WRITE()
	return inst
}

// GetSourceAccount gets the "source" account.
// The source account.
func (inst *FlashBorrowReserveLiquidity) GetSourceAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[0]
}

// SetDestinationAccount sets the "destination" account.
// The destination account.
func (inst *FlashBorrowReserveLiquidity) SetDestinationAccount(destination ag_solanago.PublicKey) *FlashBorrowReserveLiquidity {
	inst.Accounts[1] = ag_solanago.Meta(destination).WRITE()
	return inst
}

// GetDestinationAccount gets the "destination" account.
// The destination account.
func (inst *FlashBorrowReserveLiquidity) GetDestinationAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[1]
}

// SetReserveAccount sets the "reserve" account.
// The reserve account.
func (inst *FlashBorrowReserveLiquidity) SetReserveAccount(reserve ag_solanago.PublicKey) *FlashBorrowReserveLiquidity {
	inst.Accounts[2] = ag_solanago.Meta(reserve).WRITE()
	return inst
}

// GetReserveAccount gets the "reserve" account.
// The reserve account.
func (inst *FlashBorrowReserveLiquidity) GetReserveAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[2]
}

// SetLendingMarketAccount sets the "lending_market" account.
// The lending market account (read-only).
func (inst *FlashBorrowReserveLiquidity) SetLendingMarketAccount(lendingMarket ag_solanago.PublicKey) *FlashBorrowReserveLiquidity {
	inst.Accounts[3] = ag_solanago.Meta(lendingMarket)
	return inst
}

// GetLendingMarketAccount gets the "lending_market" account.
// The lending market account (read-only).
func (inst *FlashBorrowReserveLiquidity) GetLendingMarketAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[3]
}

// SetLendingMarketAuthorityAccount sets the "lending_market_authority" account.
// The lending market authority account (read-only).
func (inst *FlashBorrowReserveLiquidity) SetLendingMarketAuthorityAccount(lendingMarketAuthority ag_solanago.PublicKey) *FlashBorrowReserveLiquidity {
	inst.Accounts[4] = ag_solanago.Meta(lendingMarketAuthority)
	return inst
}

// GetLendingMarketAuthorityAccount gets the "lending_market_authority" account.
// The lending market authority account (read-only).
func (inst *FlashBorrowReserveLiquidity) GetLendingMarketAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[4]
}

// SetSysvarInstructionsAccount sets the "sysvar_instructions" account.
// The sysvar instructions account (read-only).
func (inst *FlashBorrowReserveLiquidity) SetSysvarInstructionsAccount(sysvarInstructions ag_solanago.PublicKey) *FlashBorrowReserveLiquidity {
	inst.Accounts[5] = ag_solanago.Meta(sysvarInstructions)
	return inst
}

// GetSysvarInstructionsAccount gets the "sysvar_instructions" account.
// The sysvar instructions account (read-only).
func (inst *FlashBorrowReserveLiquidity) GetSysvarInstructionsAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[5]
}

// SetTokenProgramAccount sets the "token_program" account.
// The SPL Token program account (read-only).
func (inst *FlashBorrowReserveLiquidity) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *FlashBorrowReserveLiquidity {
	inst.Accounts[6] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "token_program" account.
// The SPL Token program account (read-only).
func (inst *FlashBorrowReserveLiquidity) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.Accounts[6]
}

func (inst FlashBorrowReserveLiquidity) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_FlashBorrowReserveLiquidity),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst FlashBorrowReserveLiquidity) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *FlashBorrowReserveLiquidity) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Amount == nil {
			return errors.New("Amount parameter is not set")
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
			return fmt.Errorf("accounts.Reserve is not set")
		}
		if inst.Accounts[3] == nil {
			return fmt.Errorf("accounts.LendingMarket is not set")
		}
		if inst.Accounts[4] == nil {
			return fmt.Errorf("accounts.LendingMarketAuthority is not set")
		}
		if inst.Accounts[5] == nil {
			return fmt.Errorf("accounts.SysvarInstructions is not set")
		}
		if inst.Accounts[6] == nil {
			return fmt.Errorf("accounts.TokenProgram is not set")
		}
	}
	return nil
}

func (inst *FlashBorrowReserveLiquidity) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("FlashBorrowReserveLiquidity")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Amount", *inst.Amount))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("source_liquidity", inst.Accounts[0]))
						accountsBranch.Child(ag_format.Meta("destination_liquidity", inst.Accounts[1]))
						accountsBranch.Child(ag_format.Meta("reserve", inst.Accounts[2]))
						accountsBranch.Child(ag_format.Meta("lending_market", inst.Accounts[3]))
						accountsBranch.Child(ag_format.Meta("lending_market_authority", inst.Accounts[4]))
						accountsBranch.Child(ag_format.Meta("sysvar_instructions", inst.Accounts[5]))
						accountsBranch.Child(ag_format.Meta("token_program", inst.Accounts[6]))
					})
				})
		})
}

func (inst FlashBorrowReserveLiquidity) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	// Serialize `Amount` param:
	err := encoder.Encode(inst.Amount)
	if err != nil {
		return err
	}
	return nil
}

func (inst *FlashBorrowReserveLiquidity) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	// Deserialize `Amount`:
	err := decoder.Decode(&inst.Amount)
	if err != nil {
		return err
	}
	return nil
}

func NewFlashBorrowReserveLiquidityInstruction(
	amount uint64,
	sourceLiquidity ag_solanago.PublicKey,
	destinationLiquidity ag_solanago.PublicKey,
	reserve ag_solanago.PublicKey,
	lendingMarket ag_solanago.PublicKey,
	lendingMarketAuthority ag_solanago.PublicKey,
	sysvarInstructions ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey,
) *FlashBorrowReserveLiquidity {
	return NewFlashBorrowReserveLiquidityInstructionBuilder().
		SetAmount(amount).
		SetSourceAccount(sourceLiquidity).
		SetDestinationAccount(destinationLiquidity).
		SetReserveAccount(reserve).
		SetLendingMarketAccount(lendingMarket).
		SetLendingMarketAuthorityAccount(lendingMarketAuthority).
		SetSysvarInstructionsAccount(sysvarInstructions).
		SetTokenProgramAccount(tokenProgram)
}
