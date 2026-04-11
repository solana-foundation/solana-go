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

package computebudget

import (
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

type RequestUnitsDeprecated struct {
	// Units to request
	Units uint32

	// Additional fee to add
	AdditionalFee uint32
}

func (obj *RequestUnitsDeprecated) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	return nil
}

func (slice RequestUnitsDeprecated) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	return
}

// NewRequestUnitsDeprecatedInstructionBuilder creates a new `RequestUnitsDeprecated` instruction builder.
func NewRequestUnitsDeprecatedInstructionBuilder() *RequestUnitsDeprecated {
	nd := &RequestUnitsDeprecated{}
	return nd
}

// Units to request
func (inst *RequestUnitsDeprecated) SetUnits(units uint32) *RequestUnitsDeprecated {
	inst.Units = units
	return inst
}

// Additional fee to add
func (inst *RequestUnitsDeprecated) SetAdditionalFee(additionalFee uint32) *RequestUnitsDeprecated {
	inst.AdditionalFee = additionalFee
	return inst
}

func (inst RequestUnitsDeprecated) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_RequestUnitsDeprecated),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst RequestUnitsDeprecated) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *RequestUnitsDeprecated) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Units == 0 {
			return errors.New("units parameter is not set")
		}
		if inst.Units > MAX_COMPUTE_UNIT_LIMIT {
			return errors.New("units parameter exceeds the maximum compute unit")
		}
		if inst.AdditionalFee == 0 {
			return errors.New("AdditionalFee parameter is not set")
		}
	}
	return nil
}

func (inst *RequestUnitsDeprecated) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("RequestUnitsDeprecated")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("        Units", inst.Units))
						paramsBranch.Child(ag_format.Param("AdditionalFee", inst.AdditionalFee))
					})
				})
		})
}

// NewRequestUnitsDeprecatedInstruction declares a new RequestUnitsDeprecated instruction with the provided parameters and accounts.
func NewRequestUnitsDeprecatedInstruction(
	// Parameters:
	units uint32,
	additionalFee uint32,
) *RequestUnitsDeprecated {
	return NewRequestUnitsDeprecatedInstructionBuilder().
		SetUnits(units).
		SetAdditionalFee(additionalFee)
}
