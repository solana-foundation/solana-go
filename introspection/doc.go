// Copyright 2026 github.com/gagliardetto
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

// Package introspection resolves every instruction of a confirmed
// transaction, top-level and inner (CPI), to its program, accounts and data,
// in the order an explorer shows them.
//
// The resolved instructions implement solana.Instruction, so they can be
// decoded with solana.DecodeInstruction or with the DecodeInstruction
// function of a program package:
//
//	res, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
//		Encoding:                       solana.EncodingBase64,
//		MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion1,
//	})
//	if err != nil {
//		return err
//	}
//	ixs, err := introspection.WalkGetTransactionResult(res)
//	if err != nil {
//		return err
//	}
//	for _, ix := range ixs {
//		if !ix.ProgID.Equals(token.ProgramID) {
//			continue
//		}
//		inst, err := token.DecodeInstruction(ix.AccountValues, ix.DataBytes)
//		if err != nil {
//			return err
//		}
//		if _, ok := inst.Impl.(*token.SyncNative); ok {
//			fmt.Println("SyncNative at", ix.Trace) // e.g. "outer[1]" or "inner[2/0]"
//		}
//	}
//
// The accounts that v0 transactions load from address lookup tables are
// taken from the transaction meta (meta.loadedAddresses), so no lookup table
// has to be fetched.
//
// Legacy, v0 and v1 transactions are supported, fetched with the json,
// base64 or base64+zstd encoding, or with base58, which the RPC serves only
// for legacy and v0 transactions. The jsonParsed encoding is not supported:
// its instructions come pre-parsed, without their raw data.
//
// This package inspects transactions returned by the RPC. It is unrelated to
// the on-chain Instructions sysvar that programs use to inspect the
// transaction they run in.
package introspection
