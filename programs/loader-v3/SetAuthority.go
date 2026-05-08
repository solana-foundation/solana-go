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

// SetAuthority changes the authority on a buffer or programdata account.
// The unchecked variant is deprecated upstream in favor of
// SetAuthorityChecked when setting a new non-nil authority, but it remains
// the only way to clear an upgrade authority (pass nil for the new authority).
//
// Account references (buffer form):
//
//	[0] = [WRITE]     Buffer account
//	[1] = [SIGNER]    Current authority
//	[2] = [optional]  New authority (omit to drop the authority)
//
// Account references (programdata form):
//
//	[0] = [WRITE]     ProgramData (PDA)
//	[1] = [SIGNER]    Current authority
//	[2] = [optional]  New authority (omit to make the program immutable)
type SetAuthority struct {
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewSetAuthorityInstructionBuilder() *SetAuthority {
	return &SetAuthority{}
}

func (inst SetAuthority) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_SetAuthority, binary.LittleEndian),
	}}
}

func (inst SetAuthority) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *SetAuthority) Validate() error {
	if len(inst.AccountMetaSlice) < 2 || len(inst.AccountMetaSlice) > 3 {
		return fmt.Errorf("SetAuthority expects 2 or 3 accounts, got %d", len(inst.AccountMetaSlice))
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *SetAuthority) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("SetAuthority")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						for i, acc := range inst.AccountMetaSlice {
							a.Child(ag_format.Meta(fmt.Sprintf("Account[%d]", i), acc))
						}
					})
				})
		})
}

func (inst SetAuthority) MarshalWithEncoder(_ *ag_binary.Encoder) error    { return nil }
func (inst *SetAuthority) UnmarshalWithDecoder(_ *ag_binary.Decoder) error { return nil }

// NewSetBufferAuthorityInstruction builds a SetAuthority that transfers a
// buffer's authority. Upstream deprecates this in favor of
// NewSetBufferAuthorityCheckedInstruction, but it is retained for decoding
// historical transactions.
func NewSetBufferAuthorityInstruction(
	buffer, currentAuthority, newAuthority ag_solanago.PublicKey,
) *SetAuthority {
	inst := NewSetAuthorityInstructionBuilder()
	inst.AccountMetaSlice = ag_solanago.AccountMetaSlice{
		ag_solanago.Meta(buffer).WRITE(),
		ag_solanago.Meta(currentAuthority).SIGNER(),
		ag_solanago.Meta(newAuthority),
	}
	return inst
}

// NewSetUpgradeAuthorityInstruction builds a SetAuthority that transfers or
// drops a program's upgrade authority. Pass nil for newAuthority to make the
// program immutable.
func NewSetUpgradeAuthorityInstruction(
	program, currentAuthority ag_solanago.PublicKey,
	newAuthority *ag_solanago.PublicKey,
) *SetAuthority {
	programDataPDA := MustGetProgramDataAddress(program)
	inst := NewSetAuthorityInstructionBuilder()
	metas := ag_solanago.AccountMetaSlice{
		ag_solanago.Meta(programDataPDA).WRITE(),
		ag_solanago.Meta(currentAuthority).SIGNER(),
	}
	if newAuthority != nil {
		metas = append(metas, ag_solanago.Meta(*newAuthority))
	}
	inst.AccountMetaSlice = metas
	return inst
}
