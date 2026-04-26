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

package system

import (
	"encoding/base64"
	"testing"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	nonceAccountBase64Data := "AAAAAAEAAABHaauXIEuoP7DK7hf3ho8eB05SFYGg2J2UN52qZbcXsnM+zs3rCNyHGAjze1Gvfq4gRzzrz7ggv4rYXkMo8P2DiBMAAAAAAAA="

	decoded, err := base64.StdEncoding.DecodeString(nonceAccountBase64Data)
	assert.NoError(t, err)

	dec := bin.NewBinDecoder(decoded)

	acc := new(NonceAccount)

	err = acc.UnmarshalWithDecoder(dec)
	assert.NoError(t, err)

	assert.Equal(t, uint32(0), acc.Version)
	assert.Equal(t, uint32(1), acc.State)
	assert.Equal(t, solana.MustPublicKeyFromBase58("5omQJtDUHA3gMFdHEQg1zZSvcBUVzey5WaKWYRmqF1Vj"), acc.AuthorizedPubkey)
	assert.Equal(t, solana.MustPublicKeyFromBase58("8ksS6xXd7vzNrpZfBTf9gJ87Bma5AjnQ9baEcT7xH5QE"), acc.Nonce)
	assert.Equal(t, uint64(5000), acc.FeeCalculator.LamportsPerSignature)
}
