package token2022

import (
	"encoding/binary"
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// InterestBearingMint sub-instruction IDs.
const (
	InterestBearingMint_Initialize uint8 = iota
	InterestBearingMint_UpdateRate
)

// InterestBearingMintExtension is the instruction wrapper for the InterestBearingMint extension (ID 33).
type InterestBearingMintExtension struct {
	SubInstruction uint8

	// For Initialize: the rate authority.
	RateAuthority *ag_solanago.PublicKey `bin:"-"`
	// Rate in basis points.
	Rate int16 `bin:"-"`

	// For Initialize:
	// [0] = [WRITE] mint
	//
	// For UpdateRate:
	// [0] = [WRITE] mint
	// [1] = [] rateAuthority
	// [2...] = [SIGNER] signers
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *InterestBearingMintExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	if obj.SubInstruction == InterestBearingMint_Initialize {
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	} else {
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	}
	return nil
}

func (slice InterestBearingMintExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst InterestBearingMintExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_InterestBearingMintExtension),
	}}
}

func (inst InterestBearingMintExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *InterestBearingMintExtension) Validate() error {
	if len(inst.Accounts) == 0 || inst.Accounts[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	if inst.SubInstruction == InterestBearingMint_UpdateRate {
		if len(inst.Accounts) < 2 || inst.Accounts[1] == nil {
			return errors.New("accounts.RateAuthority is not set")
		}
		if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	}
	return nil
}

func (inst *InterestBearingMintExtension) EncodeToTree(parent ag_treeout.Branches) {
	name := "Initialize"
	if inst.SubInstruction == InterestBearingMint_UpdateRate {
		name = "UpdateRate"
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("InterestBearingMint." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						if inst.SubInstruction == InterestBearingMint_Initialize {
							paramsBranch.Child(ag_format.Param("RateAuthority (OPT)", inst.RateAuthority))
						}
						paramsBranch.Child(ag_format.Param("Rate", inst.Rate))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("         mint", inst.Accounts[0]))
						if inst.SubInstruction == InterestBearingMint_UpdateRate && len(inst.Accounts) > 1 {
							accountsBranch.Child(ag_format.Meta("rateAuthority", inst.Accounts[1]))
						}
					})
				})
		})
}

func (obj InterestBearingMintExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.WriteUint8(obj.SubInstruction)
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case InterestBearingMint_Initialize:
		if err = writeOptionalPubkey(encoder, obj.RateAuthority); err != nil {
			return err
		}
		err = encoder.WriteInt16(obj.Rate, binary.LittleEndian)
		if err != nil {
			return err
		}
	case InterestBearingMint_UpdateRate:
		err = encoder.WriteInt16(obj.Rate, binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *InterestBearingMintExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case InterestBearingMint_Initialize:
		obj.RateAuthority, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
		obj.Rate, err = decoder.ReadInt16(binary.LittleEndian)
		if err != nil {
			return err
		}
	case InterestBearingMint_UpdateRate:
		obj.Rate, err = decoder.ReadInt16(binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewInitializeInterestBearingMintInstruction creates an instruction to initialize interest-bearing mint.
func NewInitializeInterestBearingMintInstruction(
	rateAuthority *ag_solanago.PublicKey,
	rate int16,
	mint ag_solanago.PublicKey,
) *InterestBearingMintExtension {
	inst := &InterestBearingMintExtension{
		SubInstruction: InterestBearingMint_Initialize,
		RateAuthority:  rateAuthority,
		Rate:           rate,
		Accounts:       make(ag_solanago.AccountMetaSlice, 1),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// NewUpdateInterestRateInstruction creates an instruction to update the interest rate.
func NewUpdateInterestRateInstruction(
	rate int16,
	mint ag_solanago.PublicKey,
	rateAuthority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *InterestBearingMintExtension {
	inst := &InterestBearingMintExtension{
		SubInstruction: InterestBearingMint_UpdateRate,
		Rate:           rate,
		Accounts:       make(ag_solanago.AccountMetaSlice, 2),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	inst.Accounts[1] = ag_solanago.Meta(rateAuthority)
	if len(multisigSigners) == 0 {
		inst.Accounts[1].SIGNER()
	}
	for _, signer := range multisigSigners {
		inst.Signers = append(inst.Signers, ag_solanago.Meta(signer).SIGNER())
	}
	return inst
}
