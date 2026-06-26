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

package sysvar

import (
	"encoding/binary"
	"fmt"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
)

// EpochRewardsSize is the serialized size, in bytes, of the EpochRewards sysvar.
const EpochRewardsSize = 81

// EpochRewards is the data of the EpochRewards sysvar (account
// solana.SysVarEpochRewardsPubkey): the state of the partitioned epoch-rewards
// distribution. Mirrors solana-sdk epoch_rewards::EpochRewards.
type EpochRewards struct {
	// DistributionStartingBlockHeight is the block height at which rewards
	// distribution started for the current epoch.
	DistributionStartingBlockHeight uint64
	// NumPartitions is the number of partitions rewards are split across.
	NumPartitions uint64
	// ParentBlockhash is the blockhash of the epoch's last block, used to seed
	// the partition shuffle.
	ParentBlockhash solana.Hash
	// TotalPoints is the total rewards points (u128) calculated for the epoch.
	TotalPoints bin.Uint128
	// TotalRewards is the total rewards, in lamports, for the epoch.
	TotalRewards uint64
	// DistributedRewards is the rewards, in lamports, distributed so far.
	DistributedRewards uint64
	// Active reports whether rewards distribution is in progress.
	Active bool
}

// Distribute records that amount more lamports of rewards have been distributed.
// It returns an error if that would push DistributedRewards above TotalRewards
// (or overflow), mirroring the assertion in EpochRewards::distribute.
func (e *EpochRewards) Distribute(amount uint64) error {
	newDistributed := e.DistributedRewards + amount
	if newDistributed < e.DistributedRewards || newDistributed > e.TotalRewards {
		return fmt.Errorf("epoch rewards: distributing %d would exceed total rewards %d", amount, e.TotalRewards)
	}
	e.DistributedRewards = newDistributed
	return nil
}

func (e EpochRewards) MarshalWithEncoder(encoder *bin.Encoder) error {
	if err := encoder.WriteUint64(e.DistributionStartingBlockHeight, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(e.NumPartitions, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteBytes(e.ParentBlockhash[:], false); err != nil {
		return err
	}
	if err := encoder.WriteUint128(e.TotalPoints, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(e.TotalRewards, binary.LittleEndian); err != nil {
		return err
	}
	if err := encoder.WriteUint64(e.DistributedRewards, binary.LittleEndian); err != nil {
		return err
	}
	return encoder.WriteBool(e.Active)
}

func (e *EpochRewards) UnmarshalWithDecoder(decoder *bin.Decoder) error {
	var err error
	if e.DistributionStartingBlockHeight, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	if e.NumPartitions, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	buf, err := decoder.ReadNBytes(32)
	if err != nil {
		return err
	}
	copy(e.ParentBlockhash[:], buf)
	if e.TotalPoints, err = decoder.ReadUint128(binary.LittleEndian); err != nil {
		return err
	}
	if e.TotalRewards, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	if e.DistributedRewards, err = decoder.ReadUint64(binary.LittleEndian); err != nil {
		return err
	}
	e.Active, err = decoder.ReadBool()
	return err
}

func (e EpochRewards) MarshalBinary() ([]byte, error) { return encodeSysvar(e) }
func (e *EpochRewards) UnmarshalBinary(data []byte) error {
	return e.UnmarshalWithDecoder(bin.NewBinDecoder(data))
}

// DecodeEpochRewards decodes EpochRewards sysvar account data.
func DecodeEpochRewards(data []byte) (*EpochRewards, error) {
	var e EpochRewards
	if err := e.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &e, nil
}
