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
	"encoding/binary"
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	ag_binary "github.com/gagliardetto/solana-go/binary"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

const MAX_HEAP_FRAME_BYTES uint32 = 256 * 1024

type RequestHeapFrame struct {
	HeapSize uint32
}

func (obj *RequestHeapFrame) SetAccounts(accounts []*ag_solanago.AccountMeta) error {
	return nil
}

func (slice RequestHeapFrame) GetAccounts() (accounts []*ag_solanago.AccountMeta) {
	return
}

// NewRequestHeapFrameInstructionBuilder creates a new `RequestHeapFrame` instruction builder.
func NewRequestHeapFrameInstructionBuilder() *RequestHeapFrame {
	nd := &RequestHeapFrame{}
	return nd
}

// Request heap frame in bytes
func (inst *RequestHeapFrame) SetHeapSize(heapSize uint32) *RequestHeapFrame {
	inst.HeapSize = heapSize
	return inst
}

func (inst RequestHeapFrame) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint8(Instruction_RequestHeapFrame),
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst RequestHeapFrame) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *RequestHeapFrame) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.HeapSize == 0 {
			return errors.New("HeapSize parameter is not set")
		}
		if inst.HeapSize > MAX_HEAP_FRAME_BYTES {
			return errors.New("HeapSize parameter exceeds the maximum heap frame bytes")
		}
	}
	return nil
}

func (inst *RequestHeapFrame) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("RequestHeapFrame")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("HeapSize", inst.HeapSize))
					})
				})
		})
}

func (obj RequestHeapFrame) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `HeapSize` param:
	err = encoder.WriteUint32(obj.HeapSize, binary.LittleEndian)
	if err != nil {
		return err
	}
	return nil
}
func (obj *RequestHeapFrame) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `HeapSize`:
	err = decoder.Decode(&obj.HeapSize)
	if err != nil {
		return err
	}
	return nil
}

// NewRequestHeapFrameInstruction declares a new RequestHeapFrame instruction with the provided parameters and accounts.
func NewRequestHeapFrameInstruction(
	// Parameters:
	heapSize uint32,
) *RequestHeapFrame {
	return NewRequestHeapFrameInstructionBuilder().SetHeapSize(heapSize)
}
