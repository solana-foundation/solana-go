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

package cmd

import (
	"context"
	"fmt"
	"strconv"

	json "github.com/goccy/go-json"
	"github.com/spf13/cobra"
)

var getBlockCmd = &cobra.Command{
	Use:   "block {block_num}",
	Short: "Retrieve a block, with all of its transactions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		client := getClient()
		ctx := context.Background()

		var slot int64
		if slot, err = strconv.ParseInt(args[0], 10, 64); err != nil {
			return fmt.Errorf("unable to parse provided slot number %q: %w", args[0], err)
		}

		resp, err := client.GetBlock(ctx, uint64(slot))
		if err != nil {
			return err
		}

		cnt, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(cnt))

		return nil
	},
}

func init() {
	getCmd.AddCommand(getBlockCmd)
}
