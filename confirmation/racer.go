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
	"sync"
)

// raceStrategies runs the signature confirmation and a lifetime strategy
// until one of them returns, then cancels the other one and waits for it.
// Kit: raceStrategies.
//
// Kit races the signature promise first, so it wins when both promises have
// settled. Here the signature result wins when it is ready by the time the
// lifetime result is picked.
func raceStrategies(ctx context.Context, signature, lifetime func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	raceCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	signatureResult := make(chan error, 1)
	lifetimeResult := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Go(func() { lifetimeResult <- lifetime(raceCtx) })
	wg.Go(func() { signatureResult <- signature(raceCtx) })

	err := pick(signatureResult, lifetimeResult)
	cancel()
	wg.Wait()
	return contextError(ctx, err)
}

// pick returns the first result. When the lifetime result comes first, a
// signature result that is ready as well takes precedence.
func pick(signatureResult, lifetimeResult <-chan error) error {
	select {
	case err := <-signatureResult:
		return err
	case err := <-lifetimeResult:
		select {
		case sigErr := <-signatureResult:
			return sigErr
		default:
			return err
		}
	}
}

// firstOf runs fns concurrently and returns the first result. It then cancels
// the others and waits for them. Kit: safeRace.
func firstOf(ctx context.Context, fns ...func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan error, len(fns))
	var wg sync.WaitGroup
	for _, fn := range fns {
		wg.Go(func() { results <- fn(ctx) })
	}
	err := <-results
	cancel()
	wg.Wait()
	return err
}
