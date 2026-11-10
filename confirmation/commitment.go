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

package confirmation

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go/rpc"
)

// normalizeCommitment returns the commitment level to wait for. An empty
// commitment means finalized, the default of the RPC server. Kit's comparator
// throws on any other value, so other values are an error here.
func normalizeCommitment(commitment rpc.CommitmentType) (rpc.CommitmentType, error) {
	switch commitment {
	case "":
		return rpc.CommitmentFinalized, nil
	case rpc.CommitmentProcessed, rpc.CommitmentConfirmed, rpc.CommitmentFinalized:
		return commitment, nil
	default:
		return "", fmt.Errorf("confirmation: unsupported commitment %q", commitment)
	}
}

// contextError returns ctx.Err() once ctx is done, and err otherwise. A
// strategy that stops because of cancellation must report the cancellation:
// its subscription then reports ws.ErrSubscriptionClosed, and RPC calls report
// the cancellation wrapped.
func contextError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return err
}

// waitDone blocks until ctx is done. It stands for a kit promise that never
// settles, so that the other side of a race decides.
func waitDone(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
