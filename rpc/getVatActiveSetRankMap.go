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

package rpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// GetVatActiveSetRankMapOpts holds the optional parameters of GetVatActiveSetRankMap.
// It follows RpcVatActiveSetRankMapConfig of Agave v4.5.0-beta.0.
type GetVatActiveSetRankMapOpts struct {
	// Identity limits the result to one validator identity. Optional.
	Identity *solana.PublicKey
	// MinContextSlot is the minimum finalized slot the node must have reached. Optional.
	MinContextSlot *uint64
}

// GetVatActiveSetRankMapResult is the result of GetVatActiveSetRankMap.
type GetVatActiveSetRankMapResult struct {
	RPCContext
	// Value is nil when the epoch that contains the slot is unavailable.
	Value *VatActiveSetRankMap `json:"value"`
}

// VatActiveSetRankMap is the Validator Admission Ticket (VAT) active-set rank map of one epoch.
// It mirrors RpcVatActiveSetRankMap of Agave v4.5.0-beta.0.
type VatActiveSetRankMap struct {
	// Epoch is the epoch in which this rank map is effective.
	Epoch uint64 `json:"epoch"`
	// TotalStake is the total active stake of the full map, in lamports.
	TotalStake uint64 `json:"totalStake"`
	// Validators are the ranked validators, ordered by rank.
	Validators []VatActiveSetRankMapEntry `json:"validators"`
}

// VatActiveSetRankMapEntry is one validator in a VatActiveSetRankMap.
// It mirrors RpcVatActiveSetRankMapEntry of Agave v4.5.0-beta.0.
type VatActiveSetRankMapEntry struct {
	// Rank is the zero-based index of the validator in Alpenglow certificate bitmaps.
	Rank uint16 `json:"rank"`
	// VotePubkey is the vote account address.
	VotePubkey solana.PublicKey `json:"votePubkey"`
	// NodePubkey is the validator identity address.
	NodePubkey solana.PublicKey `json:"nodePubkey"`
	// BLSPubkeyCompressed is the base-58 encoding of the compressed BLS public key.
	BLSPubkeyCompressed string `json:"blsPubkeyCompressed"`
	// Stake is the stake of the validator, in lamports.
	Stake uint64 `json:"stake"`
}

// UnmarshalJSON rejects a zero total stake or validator stake, as the Rust
// client does by typing these fields as NonZero<u64>.
func (m *VatActiveSetRankMap) UnmarshalJSON(data []byte) error {
	type plain VatActiveSetRankMap
	var v plain
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v.TotalStake == 0 {
		return errors.New("totalStake must be non-zero")
	}
	for i, e := range v.Validators {
		if e.Stake == 0 {
			return fmt.Errorf("validators[%d].stake must be non-zero", i)
		}
	}
	*m = VatActiveSetRankMap(v)
	return nil
}

// GetVatActiveSetRankMap returns the Validator Admission Ticket (VAT) active-set
// rank map for the epoch that contains slot.
//
// The method always reads finalized state, and there is no commitment parameter.
// Value of the result is nil when the epoch is unavailable.
//
// The identity filter keeps the rank of the validator and the total stake of the
// full map. An unknown identity gives an empty validator list.
//
// Agave serves this method since v4.5.0-beta.0. A node without it answers with
// the JSON RPC error -32601 (method not found).
func (cl *Client) GetVatActiveSetRankMap(
	ctx context.Context,
	slot uint64,
	opts *GetVatActiveSetRankMapOpts,
) (out *GetVatActiveSetRankMapResult, err error) {
	params := []any{slot}
	if opts != nil {
		obj := M{}
		if opts.Identity != nil {
			obj["identity"] = opts.Identity
		}
		if opts.MinContextSlot != nil {
			obj["minContextSlot"] = *opts.MinContextSlot
		}
		if len(obj) != 0 {
			params = append(params, obj)
		}
	}

	err = cl.rpcClient.CallForInto(ctx, &out, "getVatActiveSetRankMap", params)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, errors.New("expected a value, got null result")
	}
	return out, nil
}
