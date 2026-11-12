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
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

// TimeoutFunc returns an error once the time to wait for commitment elapsed.
type TimeoutFunc func(ctx context.Context, commitment rpc.CommitmentType) error

const (
	processedTimeout = 30 * time.Second
	defaultTimeout   = 60 * time.Second
)

// Timeout is a TimeoutFunc that returns an error wrapping ErrTimeout after 30
// seconds for the processed commitment, and after 60 seconds otherwise. Kit:
// getTimeoutPromise.
//
// Use it only when nothing tells when the transaction expires, for example
// for an airdrop, whose transaction the server builds. After a timeout, the
// transaction may still land.
func Timeout(ctx context.Context, commitment rpc.CommitmentType) error {
	commitment, err := normalizeCommitment(commitment)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	timeout := defaultTimeout
	if commitment == rpc.CommitmentProcessed {
		timeout = processedTimeout
	}
	start := time.Now()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return fmt.Errorf("%w after %s", ErrTimeout, time.Since(start))
	}
}
