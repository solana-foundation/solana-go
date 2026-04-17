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

package cli

import (
	"fmt"

	"github.com/pkg/errors"
)

func GetDecryptPassphrase() (string, error) {
	passphrase, err := GetPassword("Enter passphrase to decrypt your vault: ")
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}

	return passphrase, nil
}
func GetEncryptPassphrase() (string, error) {
	passphrase, err := GetPassword("Enter passphrase to encrypt your vault: ")
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}

	passphraseConfirm, err := GetPassword("Confirm passphrase: ")
	if err != nil {
		return "", fmt.Errorf("reading confirmation password: %w", err)
	}

	if passphrase != passphraseConfirm {
		fmt.Println()
		return "", errors.New("passphrase mismatch!")
	}
	return passphrase, nil

}
