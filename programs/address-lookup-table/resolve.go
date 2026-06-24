// Copyright 2026 github.com/solana-foundation
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

package addresslookuptable

import (
	"context"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// ResolveMessageLookupsFromRPC fetches every address lookup table referenced
// by msg.AddressTableLookups in a single getMultipleAccounts call, decodes
// the table state, and then calls msg.SetAddressTables followed by
// msg.ResolveLookups. After it returns, msg.AccountKeys contains the full
// resolved set (static keys + dynamic keys from the lookup tables).
//
// This is the one-call convenience the issue tracker has asked for since
// #262: callers fetching a versioned transaction from the chain previously
// had to (a) extract each table address, (b) issue a getAccountInfo per
// table, (c) decode each via DecodeAddressLookupTableState, (d) build the
// map for SetAddressTables, and only then (e) call ResolveLookups. The
// helper collapses (a)-(e) into one call against an existing rpc.Client.
//
// If msg has no AddressTableLookups, the helper is a no-op and returns nil
// (mirrors ResolveLookups, which is also a no-op on legacy messages).
//
// Behavior on partial fetches:
//   - If getMultipleAccounts returns a nil entry for any referenced table,
//     the helper returns an error rather than silently dropping lookups. A
//     missing table can never resolve correctly, and continuing would
//     produce a Message whose AccountKeys is short of the entries the
//     instructions are about to index into.
func ResolveMessageLookupsFromRPC(
	ctx context.Context,
	rpcClient *rpc.Client,
	msg *solana.Message,
) error {
	return ResolveMessageLookupsFromRPCWithOpts(ctx, rpcClient, msg, nil)
}

// ResolveMessageLookupsFromRPCWithOpts is identical to
// ResolveMessageLookupsFromRPC but forwards the optional
// GetMultipleAccountsOpts to the underlying RPC call so callers can pin a
// commitment level or a minContextSlot. Encoding and DataSlice fields on
// opts are honored; passing a DataSlice that truncates the table data will
// surface as a decode error below.
func ResolveMessageLookupsFromRPCWithOpts(
	ctx context.Context,
	rpcClient *rpc.Client,
	msg *solana.Message,
	opts *rpc.GetMultipleAccountsOpts,
) error {
	if msg == nil {
		return errors.New("ResolveMessageLookupsFromRPC: nil message")
	}
	if rpcClient == nil {
		return errors.New("ResolveMessageLookupsFromRPC: nil rpc client")
	}
	lookups := msg.GetAddressTableLookups()
	if len(lookups) == 0 {
		return nil
	}
	// Already-resolved messages are a no-op: SetAddressTables would reject a
	// second resolve, so short-circuit before spending an RPC round-trip.
	if msg.IsResolved() {
		return nil
	}
	tableIDs := lookups.GetTableIDs()

	var (
		resp *rpc.GetMultipleAccountsResult
		err  error
	)
	if opts == nil {
		resp, err = rpcClient.GetMultipleAccounts(ctx, tableIDs...)
	} else {
		resp, err = rpcClient.GetMultipleAccountsWithOpts(ctx, tableIDs, opts)
	}
	if err != nil {
		return fmt.Errorf("ResolveMessageLookupsFromRPC: getMultipleAccounts: %w", err)
	}
	if resp == nil || len(resp.Value) != len(tableIDs) {
		return fmt.Errorf("ResolveMessageLookupsFromRPC: expected %d account entries, got %d", len(tableIDs), accountCount(resp))
	}

	tables := make(map[solana.PublicKey]solana.PublicKeySlice, len(tableIDs))
	for i, acct := range resp.Value {
		if acct == nil {
			return fmt.Errorf("ResolveMessageLookupsFromRPC: address lookup table %s not found on chain", tableIDs[i])
		}
		state, decodeErr := DecodeAddressLookupTableState(acct.Data.GetBinary())
		if decodeErr != nil {
			return fmt.Errorf("ResolveMessageLookupsFromRPC: decode table %s: %w", tableIDs[i], decodeErr)
		}
		tables[tableIDs[i]] = state.Addresses
	}

	if err := msg.SetAddressTables(tables); err != nil {
		return fmt.Errorf("ResolveMessageLookupsFromRPC: SetAddressTables: %w", err)
	}
	if err := msg.ResolveLookups(); err != nil {
		return fmt.Errorf("ResolveMessageLookupsFromRPC: ResolveLookups: %w", err)
	}
	return nil
}

func accountCount(resp *rpc.GetMultipleAccountsResult) int {
	if resp == nil {
		return 0
	}
	return len(resp.Value)
}
