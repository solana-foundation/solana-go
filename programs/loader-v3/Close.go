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

// Close zeroes an uninitialized, buffer, or programdata account and returns
// its lamports.
//
// Account references:
//
//	[0] = [WRITE]             Account to close
//	[1] = [WRITE]             Lamports recipient
//	[2] = [SIGNER, optional]  Authority (omit for uninitialized close)
//	[3] = [WRITE, optional]   Program (required when closing programdata)
//
// Tombstone (SIMD-0432): when true, the closed account is left as a
// tombstone rather than fully reclaimed, blocking future re-use of the
// address.
type Close struct {
	Tombstone bool

	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewCloseInstructionBuilder() *Close {
	return &Close{}
}

func (inst *Close) SetTombstone(v bool) *Close { inst.Tombstone = v; return inst }

func (inst Close) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_Close, binary.LittleEndian),
	}}
}

func (inst Close) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Close) Validate() error {
	if len(inst.AccountMetaSlice) < 2 || len(inst.AccountMetaSlice) > 4 {
		return fmt.Errorf("Close expects 2-4 accounts, got %d", len(inst.AccountMetaSlice))
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *Close) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Close")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Params").ParentFunc(func(p ag_treeout.Branches) {
						p.Child(ag_format.Param("Tombstone", inst.Tombstone))
					})
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						for i, acc := range inst.AccountMetaSlice {
							a.Child(ag_format.Meta(fmt.Sprintf("Account[%d]", i), acc))
						}
					})
				})
		})
}

func (inst Close) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	return encoder.WriteBool(inst.Tombstone)
}

func (inst *Close) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	// tombstone is an OptionalTrailingBool<false> (SIMD-0432).
	v, err := readOptionalTrailingBool(decoder, false)
	if err != nil {
		return err
	}
	inst.Tombstone = v
	return nil
}

// NewCloseInstruction is the three-account shorthand: close the target with
// the provided authority co-signing. For closing programdata (which requires
// the program account), use NewCloseAnyInstruction.
func NewCloseInstruction(
	closeAddress, recipient, authority ag_solanago.PublicKey,
	tombstone bool,
) *Close {
	return NewCloseAnyInstruction(closeAddress, recipient, &authority, nil, tombstone)
}

// NewCloseAnyInstruction mirrors the upstream `close_any` helper and
// accommodates the three distinct Close shapes: uninitialized (no authority,
// no program), buffer (authority only), and programdata (authority + program).
func NewCloseAnyInstruction(
	closeAddress, recipient ag_solanago.PublicKey,
	authority *ag_solanago.PublicKey,
	program *ag_solanago.PublicKey,
	tombstone bool,
) *Close {
	inst := NewCloseInstructionBuilder().SetTombstone(tombstone)
	metas := ag_solanago.AccountMetaSlice{
		ag_solanago.Meta(closeAddress).WRITE(),
		ag_solanago.Meta(recipient).WRITE(),
	}
	if authority != nil {
		metas = append(metas, ag_solanago.Meta(*authority).SIGNER())
	}
	if program != nil {
		metas = append(metas, ag_solanago.Meta(*program).WRITE())
	}
	inst.AccountMetaSlice = metas
	return inst
}
