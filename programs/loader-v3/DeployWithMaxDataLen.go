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
	"github.com/gagliardetto/solana-go/programs/system"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// DeployWithMaxDataLen promotes a staged buffer into an executable program,
// reserving `MaxDataLen` bytes for future upgrades.
//
// Account references:
//
//	[0] = [WRITE, SIGNER] Payer
//	[1] = [WRITE]         ProgramData (PDA)
//	[2] = [WRITE]         Program account
//	[3] = [WRITE]         Source buffer
//	[4] = []              Rent sysvar
//	[5] = []              Clock sysvar
//	[6] = []              System program
//	[7] = [SIGNER]        Upgrade authority
//
// CloseBuffer (SIMD-0430): when true, the source buffer is closed and its
// lamports returned to the payer atomically with the deploy.
type DeployWithMaxDataLen struct {
	MaxDataLen  *uint64
	CloseBuffer bool

	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func NewDeployWithMaxDataLenInstructionBuilder() *DeployWithMaxDataLen {
	return &DeployWithMaxDataLen{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 8),
	}
}

func (inst *DeployWithMaxDataLen) SetMaxDataLen(v uint64) *DeployWithMaxDataLen {
	inst.MaxDataLen = &v
	return inst
}

func (inst *DeployWithMaxDataLen) SetCloseBuffer(v bool) *DeployWithMaxDataLen {
	inst.CloseBuffer = v
	return inst
}

func (inst *DeployWithMaxDataLen) SetPayer(payer ag_solanago.PublicKey) *DeployWithMaxDataLen {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

func (inst *DeployWithMaxDataLen) SetProgramDataAccount(pda ag_solanago.PublicKey) *DeployWithMaxDataLen {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(pda).WRITE()
	return inst
}

func (inst *DeployWithMaxDataLen) SetProgramAccount(program ag_solanago.PublicKey) *DeployWithMaxDataLen {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(program).WRITE()
	return inst
}

func (inst *DeployWithMaxDataLen) SetBufferAccount(buffer ag_solanago.PublicKey) *DeployWithMaxDataLen {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(buffer).WRITE()
	return inst
}

func (inst *DeployWithMaxDataLen) SetRentSysvar(rent ag_solanago.PublicKey) *DeployWithMaxDataLen {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(rent)
	return inst
}

func (inst *DeployWithMaxDataLen) SetClockSysvar(clock ag_solanago.PublicKey) *DeployWithMaxDataLen {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(clock)
	return inst
}

func (inst *DeployWithMaxDataLen) SetSystemProgram(sys ag_solanago.PublicKey) *DeployWithMaxDataLen {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(sys)
	return inst
}

func (inst *DeployWithMaxDataLen) SetUpgradeAuthority(authority ag_solanago.PublicKey) *DeployWithMaxDataLen {
	inst.AccountMetaSlice[7] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

func (inst DeployWithMaxDataLen) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.TypeIDFromUint32(Instruction_DeployWithMaxDataLen, binary.LittleEndian),
	}}
}

func (inst DeployWithMaxDataLen) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *DeployWithMaxDataLen) Validate() error {
	if inst.MaxDataLen == nil {
		return errors.New("MaxDataLen parameter is not set")
	}
	for i, acc := range inst.AccountMetaSlice {
		if acc == nil {
			return fmt.Errorf("ins.AccountMetaSlice[%v] is not set", i)
		}
	}
	return nil
}

func (inst *DeployWithMaxDataLen) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("DeployWithMaxDataLen")).
				ParentFunc(func(ib ag_treeout.Branches) {
					ib.Child("Params").ParentFunc(func(p ag_treeout.Branches) {
						p.Child(ag_format.Param(" MaxDataLen", *inst.MaxDataLen))
						p.Child(ag_format.Param("CloseBuffer", inst.CloseBuffer))
					})
					ib.Child("Accounts").ParentFunc(func(a ag_treeout.Branches) {
						a.Child(ag_format.Meta("           Payer", inst.AccountMetaSlice[0]))
						a.Child(ag_format.Meta("     ProgramData", inst.AccountMetaSlice[1]))
						a.Child(ag_format.Meta("         Program", inst.AccountMetaSlice[2]))
						a.Child(ag_format.Meta("          Buffer", inst.AccountMetaSlice[3]))
						a.Child(ag_format.Meta("            Rent", inst.AccountMetaSlice[4]))
						a.Child(ag_format.Meta("           Clock", inst.AccountMetaSlice[5]))
						a.Child(ag_format.Meta("   SystemProgram", inst.AccountMetaSlice[6]))
						a.Child(ag_format.Meta("UpgradeAuthority", inst.AccountMetaSlice[7]))
					})
				})
		})
}

func (inst DeployWithMaxDataLen) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	if err := encoder.WriteUint64(*inst.MaxDataLen, binary.LittleEndian); err != nil {
		return err
	}
	return encoder.WriteBool(inst.CloseBuffer)
}

func (inst *DeployWithMaxDataLen) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	mdl, err := decoder.ReadUint64(binary.LittleEndian)
	if err != nil {
		return err
	}
	inst.MaxDataLen = &mdl
	// close_buffer is an OptionalTrailingBool<true> (SIMD-0430).
	v, err := readOptionalTrailingBool(decoder, true)
	if err != nil {
		return err
	}
	inst.CloseBuffer = v
	return nil
}

// NewDeployWithMaxDataLenInstruction builds the DeployWithMaxDataLen
// instruction alone. Use NewDeployWithMaxProgramLenInstructions for the full
// create-program-account + deploy pair that upstream's
// `deploy_with_max_program_len` returns.
func NewDeployWithMaxDataLenInstruction(
	payer, programDataPDA, program, buffer, upgradeAuthority ag_solanago.PublicKey,
	maxDataLen uint64,
	closeBuffer bool,
) *DeployWithMaxDataLen {
	return NewDeployWithMaxDataLenInstructionBuilder().
		SetMaxDataLen(maxDataLen).
		SetCloseBuffer(closeBuffer).
		SetPayer(payer).
		SetProgramDataAccount(programDataPDA).
		SetProgramAccount(program).
		SetBufferAccount(buffer).
		SetRentSysvar(ag_solanago.SysVarRentPubkey).
		SetClockSysvar(ag_solanago.SysVarClockPubkey).
		SetSystemProgram(ag_solanago.SystemProgramID).
		SetUpgradeAuthority(upgradeAuthority)
}

// NewDeployWithMaxProgramLenInstructions mirrors the upstream
// `deploy_with_max_program_len` helper: it allocates the program account via
// the system program and then performs the DeployWithMaxDataLen call.
func NewDeployWithMaxProgramLenInstructions(
	payer, program, buffer, upgradeAuthority ag_solanago.PublicKey,
	programLamports uint64,
	maxDataLen uint64,
	closeBuffer bool,
) []ag_solanago.Instruction {
	programDataPDA := MustGetProgramDataAddress(program)
	create := system.NewCreateAccountInstruction(
		programLamports,
		uint64(SizeOfProgram),
		ProgramID,
		payer,
		program,
	).Build()
	deploy := NewDeployWithMaxDataLenInstruction(
		payer, programDataPDA, program, buffer, upgradeAuthority,
		maxDataLen, closeBuffer,
	).Build()
	return []ag_solanago.Instruction{create, deploy}
}
