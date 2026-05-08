// Copyright 2024 github.com/cordialsys
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stake

import (
	"errors"

	bin "github.com/gagliardetto/binary"
	ag_solanago "github.com/solana-foundation/solana-go/v2"
)

type Lockup struct {
	// UnixTimestamp at which this stake will allow withdrawal, unless the transaction is signed by the custodian
	UnixTimestamp *int64
	// Epoch height at which this stake will allow withdrawal, unless the transaction is signed by the custodian
	Epoch *uint64
	// Custodian signature on a transaction exempts the operation from lockup constraints
	Custodian *ag_solanago.PublicKey
}

func (lockup *Lockup) UnmarshalWithDecoder(dec *bin.Decoder) error {
	{
		err := dec.Decode(&lockup.UnixTimestamp)
		if err != nil {
			return err
		}
	}
	{
		err := dec.Decode(&lockup.Epoch)
		if err != nil {
			return err
		}
	}
	{
		err := dec.Decode(&lockup.Custodian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (lockup *Lockup) MarshalWithEncoder(encoder *bin.Encoder) error {
	{
		err := encoder.Encode(*lockup.UnixTimestamp)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.Encode(*lockup.Epoch)
		if err != nil {
			return err
		}
	}
	{
		err := encoder.Encode(*lockup.Custodian)
		if err != nil {
			return err
		}
	}
	return nil
}

func (lockup *Lockup) Validate() error {
	if lockup.Custodian == nil {
		return errors.New("custodian parameter is not set")
	}
	if lockup.Epoch == nil {
		return errors.New("epoch parameter is not set")
	}
	if lockup.UnixTimestamp == nil {
		return errors.New("unix timestamp parameter is not set")
	}
	return nil
}
