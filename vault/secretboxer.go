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

package vault

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/solana-foundation/solana-go/v2/cli"
)

type SecretBoxer interface {
	Seal(in []byte) (string, error)
	Open(in string) ([]byte, error)
	WrapType() string
}

func SecretBoxerForType(boxerType string, keypath string) (SecretBoxer, error) {
	switch boxerType {
	case "kms-gcp":
		if keypath == "" {
			return nil, errors.New("missing kms-gcp keypath")
		}
		return NewKMSGCPBoxer(keypath), nil
	case "passphrase":
		var password string
		var err error
		if envVal := os.Getenv("SLNC_GLOBAL_INSECURE_VAULT_PASSPHRASE"); envVal != "" {
			password = envVal
		} else {
			password, err = cli.GetDecryptPassphrase()
			if err != nil {
				return nil, err
			}
		}

		return NewPassphraseBoxer(password), nil
	default:
		return nil, fmt.Errorf("unknown secret boxer: %s", boxerType)
	}
}
