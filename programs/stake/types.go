// Copyright 2024 github.com/cordialsys
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

package stake

import (
	"encoding/binary"

	ag_solanago "github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

type StakeAuthorize uint32

const (
	StakeAuthorizeStaker     StakeAuthorize = 0
	StakeAuthorizeWithdrawer StakeAuthorize = 1
)

type LockupArgs struct {
	UnixTimestamp *int64
	Epoch         *uint64
	Custodian     *ag_solanago.PublicKey
}

func (args *LockupArgs) MarshalWithEncoder(encoder *bin.Encoder) error {
	// unix_timestamp: Option<i64>
	{
		if args.UnixTimestamp != nil {
			if err := encoder.WriteOption(true); err != nil {
				return err
			}
			if err := encoder.WriteInt64(*args.UnixTimestamp, binary.LittleEndian); err != nil {
				return err
			}
		} else {
			if err := encoder.WriteOption(false); err != nil {
				return err
			}
		}
	}
	// epoch: Option<u64>
	{
		if args.Epoch != nil {
			if err := encoder.WriteOption(true); err != nil {
				return err
			}
			if err := encoder.WriteUint64(*args.Epoch, binary.LittleEndian); err != nil {
				return err
			}
		} else {
			if err := encoder.WriteOption(false); err != nil {
				return err
			}
		}
	}
	// custodian: Option<Pubkey>
	{
		if args.Custodian != nil {
			if err := encoder.WriteOption(true); err != nil {
				return err
			}
			if err := encoder.WriteBytes(args.Custodian[:], false); err != nil {
				return err
			}
		} else {
			if err := encoder.WriteOption(false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (args *LockupArgs) UnmarshalWithDecoder(dec *bin.Decoder) error {
	// unix_timestamp: Option<i64>
	{
		has, err := dec.ReadOption()
		if err != nil {
			return err
		}
		if has {
			val, err := dec.ReadInt64(binary.LittleEndian)
			if err != nil {
				return err
			}
			args.UnixTimestamp = &val
		}
	}
	// epoch: Option<u64>
	{
		has, err := dec.ReadOption()
		if err != nil {
			return err
		}
		if has {
			val, err := dec.ReadUint64(binary.LittleEndian)
			if err != nil {
				return err
			}
			args.Epoch = &val
		}
	}
	// custodian: Option<Pubkey>
	{
		has, err := dec.ReadOption()
		if err != nil {
			return err
		}
		if has {
			var val ag_solanago.PublicKey
			if err := dec.Decode(&val); err != nil {
				return err
			}
			args.Custodian = &val
		}
	}
	return nil
}

type LockupCheckedArgs struct {
	UnixTimestamp *int64
	Epoch         *uint64
}

func (args *LockupCheckedArgs) MarshalWithEncoder(encoder *bin.Encoder) error {
	// unix_timestamp: Option<i64>
	{
		if args.UnixTimestamp != nil {
			if err := encoder.WriteOption(true); err != nil {
				return err
			}
			if err := encoder.WriteInt64(*args.UnixTimestamp, binary.LittleEndian); err != nil {
				return err
			}
		} else {
			if err := encoder.WriteOption(false); err != nil {
				return err
			}
		}
	}
	// epoch: Option<u64>
	{
		if args.Epoch != nil {
			if err := encoder.WriteOption(true); err != nil {
				return err
			}
			if err := encoder.WriteUint64(*args.Epoch, binary.LittleEndian); err != nil {
				return err
			}
		} else {
			if err := encoder.WriteOption(false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (args *LockupCheckedArgs) UnmarshalWithDecoder(dec *bin.Decoder) error {
	// unix_timestamp: Option<i64>
	{
		has, err := dec.ReadOption()
		if err != nil {
			return err
		}
		if has {
			val, err := dec.ReadInt64(binary.LittleEndian)
			if err != nil {
				return err
			}
			args.UnixTimestamp = &val
		}
	}
	// epoch: Option<u64>
	{
		has, err := dec.ReadOption()
		if err != nil {
			return err
		}
		if has {
			val, err := dec.ReadUint64(binary.LittleEndian)
			if err != nil {
				return err
			}
			args.Epoch = &val
		}
	}
	return nil
}

type AuthorizeWithSeedArgs struct {
	NewAuthorizedPubkey *ag_solanago.PublicKey
	StakeAuthorize      *StakeAuthorize
	AuthoritySeed       *string
	AuthorityOwner      *ag_solanago.PublicKey
}

func (args *AuthorizeWithSeedArgs) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.WriteBytes(args.NewAuthorizedPubkey[:], false)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.WriteUint32(uint32(*args.StakeAuthorize), binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.WriteRustString(*args.AuthoritySeed)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.WriteBytes(args.AuthorityOwner[:], false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (args *AuthorizeWithSeedArgs) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&args.NewAuthorizedPubkey)
		if err != nil {
			return err
		}
	}
	{
		val, err := dec.ReadUint32(binary.LittleEndian)
		if err != nil {
			return err
		}
		sa := StakeAuthorize(val)
		args.StakeAuthorize = &sa
	}
	{
		val, err := dec.ReadRustString()
		if err != nil {
			return err
		}
		args.AuthoritySeed = &val
	}
	{
		err := dec.Decode(&args.AuthorityOwner)
		if err != nil {
			return err
		}
	}
	return nil
}

type AuthorizeCheckedWithSeedArgs struct {
	StakeAuthorize *StakeAuthorize
	AuthoritySeed  *string
	AuthorityOwner *ag_solanago.PublicKey
}

func (args *AuthorizeCheckedWithSeedArgs) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.WriteUint32(uint32(*args.StakeAuthorize), binary.LittleEndian)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.WriteRustString(*args.AuthoritySeed)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.WriteBytes(args.AuthorityOwner[:], false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (args *AuthorizeCheckedWithSeedArgs) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		val, err := dec.ReadUint32(binary.LittleEndian)
		if err != nil {
			return err
		}
		sa := StakeAuthorize(val)
		args.StakeAuthorize = &sa
	}
	{
		val, err := dec.ReadRustString()
		if err != nil {
			return err
		}
		args.AuthoritySeed = &val
	}
	{
		err := dec.Decode(&args.AuthorityOwner)
		if err != nil {
			return err
		}
	}
	return nil
}
