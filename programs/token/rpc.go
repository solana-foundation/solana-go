// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
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

package token

import (
	"context"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go/rpc"
)

const MINT_SIZE = 82

func (mint *Mint) Decode(data []byte) error {
	dec := bin.NewBinDecoder(data)
	if err := dec.Decode(&mint); err != nil {
		return fmt.Errorf("unable to decode mint: %w", err)
	}
	return nil
}

func FetchMints(ctx context.Context, rpcCli *rpc.Client) (out []*Mint, err error) {
	resp, err := rpcCli.GetProgramAccountsWithOpts(
		ctx,
		ProgramID,
		&rpc.GetProgramAccountsOpts{
			Filters: []rpc.RPCFilter{
				{
					DataSize: MINT_SIZE,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resp empty... program account not found")
	}

	for _, keyedAcct := range resp {
		acct := keyedAcct.Account

		m := new(Mint)
		if err := m.Decode(acct.Data.GetBinary()); err != nil {
			return nil, fmt.Errorf("unable to decode mint %q: %w", acct.Owner.String(), err)
		}
		out = append(out, m)

	}
	return
}
