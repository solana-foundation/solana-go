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

import solana "github.com/gagliardetto/solana-go"

// Sysvar account addresses, paired with the data types in this package. These
// re-export the canonical constants from the root package for convenience, so
// callers can fetch and decode with a single import, e.g.
//
//	resp, _ := rpcClient.GetAccountInfo(ctx, sysvar.ClockID)
//	clock, _ := sysvar.DecodeClock(resp.Value.Data.GetBinary())
var (
	ClockID           = solana.SysVarClockPubkey
	EpochRewardsID    = solana.SysVarEpochRewardsPubkey
	EpochScheduleID   = solana.SysVarEpochSchedulePubkey
	LastRestartSlotID = solana.SysVarLastRestartSlotPubkey
	RentID            = solana.SysVarRentPubkey
	SlotHashesID      = solana.SysVarSlotHashesPubkey
	SlotHistoryID     = solana.SysVarSlotHistoryPubkey
	StakeHistoryID    = solana.SysVarStakeHistoryPubkey

	// Deprecated: the Fees sysvar is deprecated; use the getFeeForMessage RPC.
	FeesID = solana.SysVarFeesPubkey
	// Deprecated: the RecentBlockhashes sysvar is deprecated; use the
	// getLatestBlockhash RPC.
	RecentBlockhashesID = solana.SysVarRecentBlockHashesPubkey
)
