// Copyright 2021 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package loaderv3

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Upgrade replaces an existing program's code with the contents of a buffer.
//
// Account references:
//
//	[0] = [WRITE]   ProgramData (PDA)
//	[1] = [WRITE]   Program account
//	[2] = [WRITE]   Source buffer
//	[3] = [WRITE]   Spill (lamports recipient)
//	[4] = []        Rent sysvar
//	[5] = []        Clock sysvar
//	[6] = [SIGNER]  Upgrade authority
//
// CloseBuffer (SIMD-0430): when true, the source buffer is closed and its
// lamports sent to the spill account atomically with the upgrade.
type Upgrade struct {
	CloseBuffer bool

	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewUpgradeInstructionBuilder() *Upgrade {
	return &Upgrade{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 7),
	}
}

func (inst *Upgrade) SetCloseBuffer(v bool) *Upgrade { inst.CloseBuffer = v; return inst }

func (inst *Upgrade) SetProgramDataAccount(pda ag_solanago.PublicKey) *Upgrade {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(pda).WRITE()
	return inst
}

func (inst *Upgrade) SetProgramAccount(program ag_solanago.PublicKey) *Upgrade {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(program).WRITE()
	return inst
}

func (inst *Upgrade) SetBufferAccount(buffer ag_solanago.PublicKey) *Upgrade {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(buffer).WRITE()
	return inst
}

func (inst *Upgrade) SetSpillAccount(spill ag_solanago.PublicKey) *Upgrade {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(spill).WRITE()
	return inst
}

func (inst *Upgrade) SetRentSysvar(rent ag_solanago.PublicKey) *Upgrade {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(rent)
	return inst
}

func (inst *Upgrade) SetClockSysvar(clock ag_solanago.PublicKey) *Upgrade {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(clock)
	return inst
}

func (inst *Upgrade) SetAuthority(authority ag_solanago.PublicKey) *Upgrade {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

func (inst Upgrade) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_Upgrade, binary.LittleEndian),
	}}
}

func (inst Upgrade) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Upgrade) Validate() error {
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *Upgrade) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Upgrade")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Params").ParentFunc(func(p ag_treeout.Branches) {
						p.Child(ag_format.Param("CloseBuffer", inst.CloseBuffer))
					})
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("ProgramData", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("    Program", inst.AccountMetaSlice[1]))
						a.Child(ag_format.Meta("     Buffer", inst.AccountMetaSlice[2]))
						a.Child(ag_format.Meta("      Spill", inst.AccountMetaSlice[3]))
						a.Child(ag_format.Meta("       Rent", inst.AccountMetaSlice[4]))
						a.Child(ag_format.Meta("      Clock", inst.AccountMetaSlice[5]))
						a.Child(ag_format.Meta("  Authority", inst.AccountMetaSlice[6]))
					})
				})
		})
}

func (inst Upgrade) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return encoder.WriteBool(inst.CloseBuffer)
}

func (inst *Upgrade) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	v, err := readOptionalTrailingBool(decoder, true)
	if err != nil {
		return err
	}
	inst.CloseBuffer = v
	return nil
}

func NewUpgradeInstruction(
	program, buffer, authority, spill ag_solanago.PublicKey,
	closeBuffer bool,
) *Upgrade {
	return NewUpgradeInstructionBuilder().
		SetCloseBuffer(closeBuffer).
		SetProgramDataAccount(MustGetProgramDataAddress(program)).
		SetProgramAccount(program).
		SetBufferAccount(buffer).
		SetSpillAccount(spill).
		SetRentSysvar(ag_solanago.SysVarRentPubkey).
		SetClockSysvar(ag_solanago.SysVarClockPubkey).
		SetAuthority(authority)
}
