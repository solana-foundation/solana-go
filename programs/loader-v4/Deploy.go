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

package loaderv4

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Deploy promotes a retracted program account into the `Deployed` state.
// The optional third account (set via NewDeployFromSourceInstruction) reads
// bytes from another program account instead of the target itself.
//
// Account references (two-account form):
//
//	[0] = [WRITE]   Program account
//	[1] = [SIGNER]  Authority
//
// Account references (deploy-from-source form):
//
//	[0] = [WRITE]   Program account
//	[1] = [SIGNER]  Authority
//	[2] = [WRITE]   Source program account
type Deploy struct {
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewDeployInstructionBuilder() *Deploy {
	return &Deploy{}
}

func (inst Deploy) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_Deploy, binary.LittleEndian),
	}}
}

func (inst Deploy) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Deploy) Validate() error {
	if n := len(inst.AccountMetaSlice); n != 2 && n != 3 {
		return fmt.Errorf("Deploy expects 2 or 3 accounts, got %d", n)
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *Deploy) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Deploy")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						for i, acc := range inst.AccountMetaSlice {
							a.Child(ag_format.Meta(fmt.Sprintf("Account[%d]", i), acc))
						}
					})
				})
		})
}

func (inst Deploy) MarshalWithEncoder(_ *ag_binary.Encoder) error    { return nil }
func (inst *Deploy) UnmarshalWithDecoder(_ *ag_binary.Decoder) error { return nil }

// NewDeployInstruction builds the two-account form: deploy the program
// using bytes already written to its own account.
func NewDeployInstruction(program, authority ag_solanago.PublicKey) *Deploy {
	inst := NewDeployInstructionBuilder()
	inst.AccountMetaSlice = ag_solanago.AccountMetaSlice{
		ag_solanago.Meta(program).WRITE(),
		ag_solanago.Meta(authority).SIGNER(),
	}
	return inst
}

// NewDeployFromSourceInstruction builds the three-account form: deploy the
// program using bytes from a separate source account, which is consumed in
// the process.
func NewDeployFromSourceInstruction(program, authority, source ag_solanago.PublicKey) *Deploy {
	inst := NewDeployInstructionBuilder()
	inst.AccountMetaSlice = ag_solanago.AccountMetaSlice{
		ag_solanago.Meta(program).WRITE(),
		ag_solanago.Meta(authority).SIGNER(),
		ag_solanago.Meta(source).WRITE(),
	}
	return inst
}
