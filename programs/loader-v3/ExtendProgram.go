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
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ExtendProgram grows a program's on-chain code buffer by `AdditionalBytes`
// (SIMD-0431 caps the minimum granularity at MINIMUM_EXTEND_PROGRAM_BYTES).
//
// Account references:
//
//	[0] = [WRITE]             ProgramData (PDA)
//	[1] = [WRITE]             Program account
//	[2] = [optional]          System program  (required when payer is provided)
//	[3] = [WRITE, SIGNER, optional] Payer (covers any additional rent)
type ExtendProgram struct {
	AdditionalBytes *uint32

	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewExtendProgramInstructionBuilder() *ExtendProgram {
	return &ExtendProgram{}
}

func (inst *ExtendProgram) SetAdditionalBytes(v uint32) *ExtendProgram {
	inst.AdditionalBytes = &v
	return inst
}

func (inst ExtendProgram) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_ExtendProgram, binary.LittleEndian),
	}}
}

func (inst ExtendProgram) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ExtendProgram) Validate() error {
	if inst.AdditionalBytes == nil {
		return errors.New("AdditionalBytes parameter is not set")
	}
	n := len(inst.AccountMetaSlice)
	if n != 2 && n != 4 {
		return fmt.Errorf("ExtendProgram expects 2 or 4 accounts, got %d", n)
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *ExtendProgram) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ExtendProgram")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Params").ParentFunc(func(p ag_treeout.Branches) {
						p.Child(ag_format.Param("AdditionalBytes", *inst.AdditionalBytes))
					})
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						for i, acc := range inst.AccountMetaSlice {
							a.Child(ag_format.Meta(fmt.Sprintf("Account[%d]", i), acc))
						}
					})
				})
		})
}

func (inst ExtendProgram) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return encoder.WriteUint32(*inst.AdditionalBytes, binary.LittleEndian)
}

func (inst *ExtendProgram) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	v, err := decoder.ReadUint32(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.AdditionalBytes = &v
	return nil
}

// NewExtendProgramInstruction builds the ExtendProgram instruction. Pass a
// nil `payer` for the two-account form (which requires the program to
// already hold enough lamports for the new size).
func NewExtendProgramInstruction(
	program ag_solanago.PublicKey,
	payer *ag_solanago.PublicKey,
	additionalBytes uint32,
) *ExtendProgram {
	programDataPDA := MustGetProgramDataAddress(program)
	inst := NewExtendProgramInstructionBuilder().SetAdditionalBytes(additionalBytes)
	metas := ag_solanago.AccountMetaSlice{
		ag_solanago.Meta(programDataPDA).WRITE(),
		ag_solanago.Meta(program).WRITE(),
	}
	if payer != nil {
		metas = append(metas,
			ag_solanago.Meta(ag_solanago.SystemProgramID),
			ag_solanago.Meta(*payer).WRITE().SIGNER(),
		)
	}
	inst.AccountMetaSlice = metas
	return inst
}
