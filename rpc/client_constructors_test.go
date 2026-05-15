// Copyright 2026 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0

package rpc

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

// Cover the constructor variants requested in #414: a default commitment can
// be pinned at construction time and read back without callers having to
// thread it through every method, and the HTTP timeout is no longer
// hard-wired to the 5-minute default.

func TestNew_DefaultCommitment_EmptyByDefault(t *testing.T) {
	cl := New("http://localhost:8899")
	if cl.DefaultCommitment() != "" {
		t.Fatalf("expected empty default commitment, got %q", cl.DefaultCommitment())
	}
}

func TestNewWithCommitment_StoresDefault(t *testing.T) {
	cl := NewWithCommitment("http://localhost:8899", CommitmentFinalized)
	if cl.DefaultCommitment() != CommitmentFinalized {
		t.Fatalf("expected %q, got %q", CommitmentFinalized, cl.DefaultCommitment())
	}
}

func TestNewWithTimeout_HonorsTimeoutOnHTTPClient(t *testing.T) {
	timeout := 17 * time.Second
	cl := NewWithTimeout("http://localhost:8899", timeout)
	httpClient := extractHTTPClient(t, cl)
	if httpClient.Timeout != timeout {
		t.Fatalf("expected http.Client.Timeout=%s, got %s", timeout, httpClient.Timeout)
	}
}

func TestNewWithTimeout_DefaultCommitmentEmpty(t *testing.T) {
	cl := NewWithTimeout("http://localhost:8899", 30*time.Second)
	if cl.DefaultCommitment() != "" {
		t.Fatalf("expected empty default commitment, got %q", cl.DefaultCommitment())
	}
}

func TestNewWithTimeoutAndCommitment_StoresBoth(t *testing.T) {
	timeout := 7 * time.Second
	cl := NewWithTimeoutAndCommitment("http://localhost:8899", timeout, CommitmentConfirmed)
	if cl.DefaultCommitment() != CommitmentConfirmed {
		t.Fatalf("expected %q, got %q", CommitmentConfirmed, cl.DefaultCommitment())
	}
	httpClient := extractHTTPClient(t, cl)
	if httpClient.Timeout != timeout {
		t.Fatalf("expected http.Client.Timeout=%s, got %s", timeout, httpClient.Timeout)
	}
}

// extractHTTPClient pulls the *http.Client back out of the underlying
// jsonrpc.RPCClient so the constructor's timeout choice can be asserted
// without exporting new fields. The reflective access is scoped to this
// test file.
func extractHTTPClient(t *testing.T, cl *Client) *http.Client {
	t.Helper()
	rpcClient := cl.rpcClient
	if rpcClient == nil {
		t.Fatal("rpc client is nil")
	}
	v := reflect.ValueOf(rpcClient)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName("httpClient")
	if !field.IsValid() {
		t.Fatalf("rpcClient %T has no httpClient field", rpcClient)
	}
	if field.IsNil() {
		t.Fatal("httpClient field is nil")
	}
	httpClient, ok := reflect.NewAt(field.Type(), field.Addr().UnsafePointer()).Elem().Interface().(*http.Client)
	if !ok {
		t.Fatalf("httpClient field is not *http.Client (%v)", field.Type())
	}
	return httpClient
}
