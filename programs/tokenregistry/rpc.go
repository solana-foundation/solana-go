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

package tokenregistry

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetTokenRegistryEntry(ctx context.Context, rpcCli *rpc.Client, mintAddress solana.PublicKey) (*TokenMeta, error) {
	resp, err := rpcCli.GetProgramAccountsWithOpts(
		ctx,
		ProgramID(),
		&rpc.GetProgramAccountsOpts{
			Filters: []rpc.RPCFilter{
				{
					Memcmp: &rpc.RPCFilterMemcmp{
						Offset: 5,
						Bytes:  mintAddress[:], // hackey to convert [32]byte to []byte
					},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resp empty... cannot find account")
	}

	if len(resp) == 0 {
		return nil, rpc.ErrNotFound
	}
	acct := resp[0].Account
	t, err := DecodeTokenMeta(acct.Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("unable to decode token meta %q: %w", acct.Owner.String(), err)
	}
	return t, nil
}

func GetEntries(ctx context.Context, rpcCli *rpc.Client) (out []*TokenMeta, err error) {
	resp, err := rpcCli.GetProgramAccountsWithOpts(
		ctx,
		ProgramID(),
		&rpc.GetProgramAccountsOpts{
			Filters: []rpc.RPCFilter{
				{
					DataSize: TOKEN_META_SIZE,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resp empty... cannot find accounts")
	}

	for _, keyedAcct := range resp {
		acct := keyedAcct.Account
		t, err := DecodeTokenMeta(acct.Data.GetBinary())
		if err != nil {
			return nil, fmt.Errorf("unable to decode token meta %q: %w", acct.Owner.String(), err)
		}
		out = append(out, t)
	}
	return out, nil
}
