package vote

import (
	"encoding/binary"
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/treeout"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/text/format"
)

// AuthorizeChecked is the checked variant of Authorize.
// Data: VoteAuthorize (u32 LE, plus BLS fields if kind=2).
type AuthorizeChecked struct {
	VoteAuthorize *VoteAuthorizeKind

	// [0] = [WRITE] VoteAccount
	// [1] = [] SysVarClock
	// [2] = [SIGNER] CurrentAuthority
	// [3] = [SIGNER] NewAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewAuthorizeCheckedInstructionBuilder() *AuthorizeChecked {
	return &AuthorizeChecked{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
}

func NewAuthorizeCheckedInstruction(
	kind VoteAuthorizeKind,
	voteAccount solana.PublicKey,
	currentAuthority solana.PublicKey,
	newAuthority solana.PublicKey,
) *AuthorizeChecked {
	return NewAuthorizeCheckedInstructionBuilder().
		SetVoteAuthorize(kind).
		SetVoteAccount(voteAccount).
		SetClockSysvar(solana.SysVarClockPubkey).
		SetCurrentAuthority(currentAuthority).
		SetNewAuthority(newAuthority)
}

func (inst *AuthorizeChecked) SetVoteAuthorize(k VoteAuthorizeKind) *AuthorizeChecked {
	inst.VoteAuthorize = &k
	return inst
}
func (inst *AuthorizeChecked) SetVoteAccount(pk solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice[0] = solana.Meta(pk).WRITE()
	return inst
}
func (inst *AuthorizeChecked) SetClockSysvar(pk solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice[1] = solana.Meta(pk)
	return inst
}
func (inst *AuthorizeChecked) SetCurrentAuthority(pk solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice[2] = solana.Meta(pk).SIGNER()
	return inst
}
func (inst *AuthorizeChecked) SetNewAuthority(pk solana.PublicKey) *AuthorizeChecked {
	inst.AccountMetaSlice[3] = solana.Meta(pk).SIGNER()
	return inst
}

func (inst AuthorizeChecked) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_AuthorizeChecked, bin.LE),
	}}
}

func (inst AuthorizeChecked) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *AuthorizeChecked) Validate() error {
	if inst.VoteAuthorize == nil {
		return errors.New("VoteAuthorize parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *AuthorizeChecked) UnmarshalWithDecoder(dec *bin.Decoder) error {
	raw, err := dec.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	k := VoteAuthorizeKind(raw)
	inst.VoteAuthorize = &k
	return nil
}

func (inst AuthorizeChecked) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.VoteAuthorize == nil {
		return errors.New("AuthorizeChecked.VoteAuthorize is nil")
	}
	return enc.WriteUint32(uint32(*inst.VoteAuthorize), binary.LittleEndian)
}

func (inst *AuthorizeChecked) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("AuthorizeChecked")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						paramsBranch.Child(format.Param("VoteAuthorize", inst.VoteAuthorize))
					})
					instructionBranch.Child("Accounts").ParentFunc(func(accountsBranch treeout.Branches) {
						accountsBranch.Child(format.Meta("     VoteAccount", inst.AccountMetaSlice[0]))
						accountsBranch.Child(format.Meta("     ClockSysvar", inst.AccountMetaSlice[1]))
						accountsBranch.Child(format.Meta("CurrentAuthority", inst.AccountMetaSlice[2]))
						accountsBranch.Child(format.Meta("    NewAuthority", inst.AccountMetaSlice[3]))
					})
				})
		})
}
