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

package main

import (
	"context"
	"fmt"

	"github.com/solana-foundation/solana-go/v2/rpc"
)

func main() {
	ctx := context.Background()
	client := rpc.New(rpc.MainNetBeta_RPC)

	out, err := client.GetLeaderSchedule(ctx)
	if err != nil {
		panic(err)
	}

	// The response maps every validator in the current epoch to its
	// scheduled slots; on mainnet this is huge. Only print a summary.
	fmt.Println("validators in current epoch:", len(out))
	count := 0
	for validator, slots := range out {
		if count >= 5 {
			break
		}
		fmt.Printf("  %s -> %d slots\n", validator, len(slots))
		count++
	}
}
