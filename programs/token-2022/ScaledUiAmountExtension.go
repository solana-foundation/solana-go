package token2022

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ScaledUiAmount sub-instruction IDs.
const (
	ScaledUiAmount_Initialize uint8 = iota
	ScaledUiAmount_UpdateMultiplier
)

// ScaledUiAmountExtension is the instruction wrapper for the ScaledUiAmount extension (ID 43).
type ScaledUiAmountExtension struct {
	SubInstruction uint8

	// The authority that can update the multiplier.
	Authority *ag_solanago.PublicKey `bin:"-"`
	// The multiplier.
	Multiplier float64 `bin:"-"`
	// For UpdateMultiplier: effective timestamp for the new multiplier.
	EffectiveTimestamp int64 `bin:"-"`

	// For Initialize:
	// [0] = [WRITE] mint
	//
	// For UpdateMultiplier:
	// [0] = [WRITE] mint
	// [1] = [] authority
	// [2...] = [SIGNER] signers
	Accounts ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
	Signers  ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (obj *ScaledUiAmountExtension) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	if obj.SubInstruction == ScaledUiAmount_Initialize {
		obj.Accounts = ag_solanago.AccountMetaSlice(accounts)
	} else {
		obj.Accounts, obj.Signers = ag_solanago.AccountMetaSlice(accounts).SplitFrom(2)
	}
	return nil
}

func (slice ScaledUiAmountExtension) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	accounts = append(accounts, slice.Accounts...)
	accounts = append(accounts, slice.Signers...)
	return
}

func (inst ScaledUiAmountExtension) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_ScaledUiAmountExtension),
	}}
}

func (inst ScaledUiAmountExtension) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ScaledUiAmountExtension) Validate() error {
	if len(inst.Accounts) == 0 || inst.Accounts[0] == nil {
		return errors.New("accounts.Mint is not set")
	}
	if inst.SubInstruction == ScaledUiAmount_UpdateMultiplier {
		if len(inst.Accounts) < 2 || inst.Accounts[1] == nil {
			return errors.New("accounts.Authority is not set")
		}
		if !inst.Accounts[1].IsSigner && len(inst.Signers) == 0 {
			return fmt.Errorf("accounts.Signers is not set")
		}
	}
	return nil
}

func (inst *ScaledUiAmountExtension) EncodeToTree(parent ag_treeout.Branches) {
	name := "Initialize"
	if inst.SubInstruction == ScaledUiAmount_UpdateMultiplier {
		name = "UpdateMultiplier"
	}
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ScaledUiAmount." + name)).
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						if inst.SubInstruction == ScaledUiAmount_Initialize {
							paramsBranch.Child(ag_format.Param("Authority (OPT)", inst.Authority))
						}
						paramsBranch.Child(ag_format.Param("Multiplier", inst.Multiplier))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("     mint", inst.Accounts[0]))
						if inst.SubInstruction == ScaledUiAmount_UpdateMultiplier && len(inst.Accounts) > 1 {
							accountsBranch.Child(ag_format.Meta("authority", inst.Accounts[1]))
						}
					})
				})
		})
}

func (obj ScaledUiAmountExtension) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	err = encoder.WriteUint8(obj.SubInstruction)
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case ScaledUiAmount_Initialize:
		if err = writeOptionalPubkey(encoder, obj.Authority); err != nil {
			return err
		}
		err = encoder.WriteUint64(math.Float64bits(obj.Multiplier), binary.LittleEndian)
		if err != nil {
			return err
		}
	case ScaledUiAmount_UpdateMultiplier:
		err = encoder.WriteUint64(math.Float64bits(obj.Multiplier), binary.LittleEndian)
		if err != nil {
			return err
		}
		err = encoder.WriteInt64(obj.EffectiveTimestamp, binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *ScaledUiAmountExtension) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	obj.SubInstruction, err = decoder.ReadUint8()
	if err != nil {
		return err
	}
	switch obj.SubInstruction {
	case ScaledUiAmount_Initialize:
		obj.Authority, err = readOptionalPubkey(decoder)
		if err != nil {
			return err
		}
		bits, err := decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		obj.Multiplier = math.Float64frombits(bits)
	case ScaledUiAmount_UpdateMultiplier:
		bits, err := decoder.ReadUint64(binary.LittleEndian)
		if err != nil {
			return err
		}
		obj.Multiplier = math.Float64frombits(bits)
		obj.EffectiveTimestamp, err = decoder.ReadInt64(binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewInitializeScaledUiAmountInstruction creates an instruction to initialize the scaled UI amount extension.
func NewInitializeScaledUiAmountInstruction(
	authority *ag_solanago.PublicKey,
	multiplier float64,
	mint ag_solanago.PublicKey,
) *ScaledUiAmountExtension {
	inst := &ScaledUiAmountExtension{
		SubInstruction: ScaledUiAmount_Initialize,
		Authority:      authority,
		Multiplier:     multiplier,
		Accounts:       make(ag_solanago.AccountMetaSlice, 1),
		Signers:        make(ag_solanago.AccountMetaSlice, 0),
	}
	inst.Accounts[0] = ag_solanago.Meta(mint).WRITE()
	return inst
}

// NewUpdateScaledUiAmountMultiplierInstruction creates an instruction to update the multiplier.
func NewUpdateScaledUiAmountMultiplierInstruction(
	multiplier float64,
	effectiveTimestamp int64,
	mint ag_solanago.PublicKey,
	authority ag_solanago.PublicKey,
	multisigSigners []ag_solanago.PublicKey,
) *ScaledUiAmountExtension {
	inst := &ScaledUiAmountExtension{
		SubInstruction:     ScaledUiAmount_UpdateMultiplier,
		Multiplier:         multiplier,
		EffectiveTimestamp: effectiveTimestamp,
		Accounts:           make(ag_solanago.AccountMetaSlice, 2),
		Signers:            make(ag_solanago.AccountMetaSlice, 0),
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
