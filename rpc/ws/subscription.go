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

package ws

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Subscription receives the notifications of one subscription. The typed
// subscriptions, such as SlotSubscription, wrap it.
type Subscription struct {
	// req is the subscribe request. It is replaced when the request is sent
	// again under a new id, always with the client lock held.
	req *request
	// subID is the server's subscription id. It is only meaningful once acked
	// is true, because 0 is a valid id.
	subID uint64
	// notificationMethod is the method of the notifications of this
	// subscription, such as slotNotification. Together with subID it keys the
	// subscription in the client, so that a server that numbers ids per
	// method cannot deliver the notifications of one method to another.
	notificationMethod string
	stream             chan result
	err                chan error
	closeFunc          func(err error)
	mu                 sync.Mutex
	closed             bool
	unsubscribeMethod  string
	decoderFunc        decoderFunc

	// subscribed is closed once, when the server confirms the subscription or
	// when the subscription closes first.
	subscribed chan struct{}
	// acked reports whether the server confirmed the subscription. It is
	// written with both the client lock and mu held.
	acked bool
	// closeErr is the error the subscription was closed with. Guarded by mu.
	closeErr error
	// staleIDs holds the server ids that were unsubscribed while this
	// subscription waited for its confirmation. Guarded by the client lock.
	staleIDs map[uint64]struct{}
}

type decoderFunc func([]byte) (any, error)

func newSubscription(
	req *request,
	closeFunc func(err error),
	unsubscribeMethod string,
	decoderFunc decoderFunc,
) *Subscription {
	return &Subscription{
		req:                req,
		subID:              0,
		notificationMethod: notificationMethod(req.Method),
		stream:             make(chan result, 200),
		err:                make(chan error, 1),
		closeFunc:          closeFunc,
		unsubscribeMethod:  unsubscribeMethod,
		decoderFunc:        decoderFunc,
		subscribed:         make(chan struct{}),
	}
}

// notificationMethod returns the method of the notifications that the
// subscribe method subscribeMethod produces: slotSubscribe gives
// slotNotification, as every method of the PubSub API does.
func notificationMethod(subscribeMethod string) string {
	return strings.TrimSuffix(subscribeMethod, "Subscribe") + "Notification"
}

// subscriptionKey identifies a confirmed subscription on the server: its
// notification method and its server subscription id.
type subscriptionKey struct {
	method string
	subID  uint64
}

func (s *Subscription) key() subscriptionKey {
	return subscriptionKey{method: s.notificationMethod, subID: s.subID}
}

// Recv returns the next notification. After the subscription closes, Recv
// first returns the notifications that are still buffered, and then an error
// on every call. That error matches ErrSubscriptionClosed with errors.Is, and
// it wraps the cause when the subscription was closed by one: a server error
// reply, a decode error, or a full buffer.
func (s *Subscription) Recv(ctx context.Context) (any, error) {
	return s.recv(ctx)
}

func (s *Subscription) recv(ctx context.Context) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-s.stream:
		if !ok {
			return nil, s.closeError()
		}
		return d, nil
	}
}

// recvAs is recv for the typed subscriptions, whose decoder returns a *T.
func recvAs[T any](ctx context.Context, s *Subscription) (*T, error) {
	d, err := s.recv(ctx)
	if err != nil {
		return nil, err
	}
	return d.(*T), nil
}

// closeError returns the error that Recv returns once the subscription is
// closed. It always matches ErrSubscriptionClosed, so loops that stop on it
// keep working, and it wraps the cause of the close when there is one.
func (s *Subscription) closeError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closeErr == nil || errors.Is(s.closeErr, ErrSubscriptionClosed) {
		return ErrSubscriptionClosed
	}
	return fmt.Errorf("%w: %w", ErrSubscriptionClosed, s.closeErr)
}

// WaitSubscribed blocks until the server confirms the subscription. It returns
// nil once the subscription is confirmed. If the subscription closes first, it
// returns an error that matches ErrSubscriptionClosed and wraps the cause, such
// as the server's error reply. If ctx is done first, it returns ctx.Err().
//
// Notifications are only sent after the server has registered the
// subscription. Wait for it before reading state that a notification would
// otherwise report, so that no change falls between the read and the
// subscription.
func (s *Subscription) WaitSubscribed(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.subscribed:
	}
	s.mu.Lock()
	if s.acked {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()
	return s.closeError()
}

// Unsubscribe closes the subscription. Recv then returns the notifications
// that are still buffered, followed by ErrSubscriptionClosed. The server side
// is unsubscribed once no other subscription of the client shares its server
// id. Calling Unsubscribe again has no effect.
func (s *Subscription) Unsubscribe() {
	s.closeFunc(ErrSubscriptionClosed)
}

// markSubscribed records the server's confirmation. It must be called with the
// client lock held, and it reports false if the subscription was already
// confirmed or closed.
func (s *Subscription) markSubscribed(subID uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.acked {
		return false
	}
	s.subID = subID
	s.acked = true
	close(s.subscribed)
	return true
}

// close closes the subscription with err and releases everything that waits
// on it. It is a no-op if the subscription is already closed.
func (s *Subscription) close(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.err <- err:
	default:
	}
	s.closeErr = err
	s.closed = true
	close(s.stream)
	close(s.err)
	if !s.acked {
		close(s.subscribed)
	}
}
