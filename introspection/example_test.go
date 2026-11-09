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

package introspection_test

import (
	"fmt"

	"github.com/gagliardetto/solana-go/introspection"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

// loadResult reads a saved getTransaction result. In a program, it comes from
// the RPC:
//
//	res, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
//		Encoding:                       solana.EncodingBase64,
//		MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion1,
//	})
func loadResult(path string) *rpc.GetTransactionResult {
	res, err := readResult(path)
	if err != nil {
		panic(err)
	}
	return res
}

// Find the Token program SyncNative instructions of a transaction, whether
// top-level or emitted by a CPI.
func ExampleWalkGetTransactionResult() {
	res := loadResult("testdata/sync_native.base64.json")

	ixs, err := introspection.WalkGetTransactionResult(res)
	if err != nil {
		panic(err)
	}
	for _, ix := range ixs {
		if !ix.ProgID.Equals(token.ProgramID) {
			continue
		}
		inst, err := token.DecodeInstruction(ix.AccountValues, ix.DataBytes)
		if err != nil {
			panic(err)
		}
		if _, ok := inst.Impl.(*token.SyncNative); ok {
			fmt.Println("SyncNative at", ix.Trace)
		}
	}
	// Output: SyncNative at outer[1]
}

// List every instruction of a v1 swap in the order an explorer shows them:
// the inner instructions right after the top-level instruction whose CPIs
// emitted them.
func ExampleWalk() {
	res := loadResult("testdata/swap_v1.json.json")
	tx, err := res.Transaction.GetTransaction()
	if err != nil {
		panic(err)
	}

	ixs, err := introspection.Walk(&tx.Message, res.Meta)
	if err != nil {
		panic(err)
	}
	for _, ix := range ixs {
		fmt.Printf("%-10s %s\n", ix.Trace, ix.ProgID)
	}
	// Output:
	// outer[0]   JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4
	// outer[1]   DF1ow4tspfHX9JwWJsAb9epbkA8hmpSEAtxXy1V27QBH
	// inner[1/0] BiSoNHVpsVZW2F7rx2eQ59yQwKxzU5NvBcmKshCSUypi
	// inner[1/1] TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA
	// inner[1/2] TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA
	// inner[1/3] DF1ow4tspfHX9JwWJsAb9epbkA8hmpSEAtxXy1V27QBH
}
