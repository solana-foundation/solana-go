package vote

import (
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/text/format"
	"github.com/gagliardetto/treeout"
)

// AuthorizeWithSeed authorizes using a derived key.
// Data: VoteAuthorizeWithSeedArgs.
type AuthorizeWithSeed struct {
	Args *VoteAuthorizeWithSeedArgs

	// [0] = [WRITE] VoteAccount
	// [1] = [] SysVarClock
	// [2] = [SIGNER] AuthorityBase
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewAuthorizeWithSeedInstructionBuilder() *AuthorizeWithSeed {
	return &AuthorizeWithSeed{
		AccountMetaSlice: make(solana.AccountMetaSlice, 3),
	}
}

func NewAuthorizeWithSeedInstruction(
	args VoteAuthorizeWithSeedArgs,
	voteAccount solana.PublicKey,
	authorityBase solana.PublicKey,
) *AuthorizeWithSeed {
	inst := NewAuthorizeWithSeedInstructionBuilder()
	inst.Args = &args
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(solana.SysVarClockPubkey)
	inst.AccountMetaSlice[2] = solana.Meta(authorityBase).SIGNER()
	return inst
}

func (inst AuthorizeWithSeed) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_AuthorizeWithSeed, bin.LE),
	}}
}

func (inst AuthorizeWithSeed) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *AuthorizeWithSeed) Validate() error {
	if inst.Args == nil {
		return errors.New("args parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *AuthorizeWithSeed) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Args = new(VoteAuthorizeWithSeedArgs)
	return inst.Args.UnmarshalWithDecoder(dec)
}

func (inst AuthorizeWithSeed) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Args == nil {
		return errors.New("AuthorizeWithSeed.Args is nil")
	}
	return inst.Args.MarshalWithEncoder(enc)
}

func (inst *AuthorizeWithSeed) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("AuthorizeWithSeed")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if inst.Args != nil {
							paramsBranch.Child(format.Param("Seed", inst.Args.CurrentAuthorityDerivedKeySeed))
						}
					})
				})
		})
}

// AuthorizeCheckedWithSeed is the checked variant of AuthorizeWithSeed.
// The new authority is also passed as a signer.
// Data: VoteAuthorizeCheckedWithSeedArgs.
type AuthorizeCheckedWithSeed struct {
	Args *VoteAuthorizeCheckedWithSeedArgs

	// [0] = [WRITE] VoteAccount
	// [1] = [] SysVarClock
	// [2] = [SIGNER] AuthorityBase
	// [3] = [SIGNER] NewAuthority
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewAuthorizeCheckedWithSeedInstructionBuilder() *AuthorizeCheckedWithSeed {
	return &AuthorizeCheckedWithSeed{
		AccountMetaSlice: make(solana.AccountMetaSlice, 4),
	}
}

func NewAuthorizeCheckedWithSeedInstruction(
	args VoteAuthorizeCheckedWithSeedArgs,
	voteAccount solana.PublicKey,
	authorityBase solana.PublicKey,
	newAuthority solana.PublicKey,
) *AuthorizeCheckedWithSeed {
	inst := NewAuthorizeCheckedWithSeedInstructionBuilder()
	inst.Args = &args
	inst.AccountMetaSlice[0] = solana.Meta(voteAccount).WRITE()
	inst.AccountMetaSlice[1] = solana.Meta(solana.SysVarClockPubkey)
	inst.AccountMetaSlice[2] = solana.Meta(authorityBase).SIGNER()
	inst.AccountMetaSlice[3] = solana.Meta(newAuthority).SIGNER()
	return inst
}

func (inst AuthorizeCheckedWithSeed) Build() *Instruction {
	return &Instruction{BaseVariant: bin.BaseVariant{
		Impl:   inst,
		TypeID: bin.TypeIDFromUint32(Instruction_AuthorizeCheckedWithSeed, bin.LE),
	}}
}

func (inst AuthorizeCheckedWithSeed) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *AuthorizeCheckedWithSeed) Validate() error {
	if inst.Args == nil {
		return errors.New("args parameter is not set")
	}
	for i, a := range inst.AccountMetaSlice {
		if a == nil {
			return fmt.Errorf("accounts[%d] is not set", i)
		}
	}
	return nil
}

func (inst *AuthorizeCheckedWithSeed) UnmarshalWithDecoder(dec *bin.Decoder) error {
	inst.Args = new(VoteAuthorizeCheckedWithSeedArgs)
	return inst.Args.UnmarshalWithDecoder(dec)
}

func (inst AuthorizeCheckedWithSeed) MarshalWithEncoder(enc *bin.Encoder) error {
	if inst.Args == nil {
		return errors.New("AuthorizeCheckedWithSeed.Args is nil")
	}
	return inst.Args.MarshalWithEncoder(enc)
}

func (inst *AuthorizeCheckedWithSeed) EncodeToTree(parent treeout.Branches) {
	parent.Child(format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch treeout.Branches) {
			programBranch.Child(format.Instruction("AuthorizeCheckedWithSeed")).
				ParentFunc(func(instructionBranch treeout.Branches) {
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch treeout.Branches) {
						if inst.Args != nil {
							paramsBranch.Child(format.Param("Seed", inst.Args.CurrentAuthorityDerivedKeySeed))
						}
					})
				})
		})
}
