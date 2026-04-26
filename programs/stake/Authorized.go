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
	"errors"

	ag_solanago "github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
)

type Authorized struct {
	// Address that will own the stake account
	Staker *ag_solanago.PublicKey
	// Address that is permitted to with from the stake account
	Withdrawer *ag_solanago.PublicKey
}

func (auth *Authorized) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&auth.Staker)
		if err != nil {
			return err
		}
	}
	{
		err := dec.Decode(&auth.Withdrawer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (auth *Authorized) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.WriteBytes(auth.Staker[:], false)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.WriteBytes(auth.Withdrawer[:], false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (auth *Authorized) Validate() error {
	if auth.Staker == nil {
		return errors.New("staker parameter is not set")
	}
	if auth.Withdrawer == nil {
		return errors.New("withdrawer parameter is not set")
	}
	return nil
}
