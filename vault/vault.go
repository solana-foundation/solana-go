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

	"github.com/solana-foundation/solana-go/v2"
)

// Vault represents a `solana-go` wallet.  It contains the encrypted
// material to load a KeyBag, which is the signing provider for
// signing transactions using the `solana-go` library.
type Vault struct {
	Kind    string `json:"kind"`
	Version int    `json:"version"`
	Comment string `json:"comment"`

	SecretBoxWrap       string `json:"secretbox_wrap"`
	SecretBoxCiphertext string `json:"secretbox_ciphertext"`

	KeyBag []solana.PrivateKey `json:"-"`
}

// NewVaultFromWalletFile returns a new Vault instance from the
// provided filename of an eos wallet.
func NewVaultFromWalletFile(filename string) (*Vault, error) {
	v := NewVault()
	fl, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fl.Close()

	err = json.NewDecoder(fl).Decode(&v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// // NewVaultFromKeysFile creates a new Vault from the keys in the
// // provided keys file.
// // keysFile should be formatted with a single private key per line
// func NewVaultFromKeysFile(keysFile string) (*Vault, error) {
// 	v := NewVault()
// 	if err := v.KeyBag.ImportFromFile(keysFile); err != nil {
// 		return nil, err
// 	}
// 	return v, nil
// }

// NewVaultFromSingleKey creates a new Vault from the provided
// private key.
func NewVaultFromSingleKey(privKey string) (*Vault, error) {
	v := NewVault()
	key, err := solana.PrivateKeyFromBase58(privKey)
	if err != nil {
		return nil, fmt.Errorf("import private key: %w", err)
	}
	v.KeyBag = append(v.KeyBag, key)
	return v, nil
}

// NewVault returns an empty vault, unsaved and with no keys.
func NewVault() *Vault {
	return &Vault{
		Kind:    "solana-vault-wallet",
		Version: 1,
	}
}

// NewKeyPair creates a new keypair, saves the private key in the
// local wallet and returns the public key. It does NOT save the
// wallet, you better do that soon after.
func (v *Vault) NewKeyPair() (pub solana.PublicKey, err error) {
	privKey, err := solana.NewRandomPrivateKey()
	if err != nil {
		return solana.PublicKey{}, err
	}

	v.KeyBag = append(v.KeyBag, privKey)

	return privKey.PublicKey(), nil
}

// AddPrivateKey appends the provided private key into the Vault's KeyBag
func (v *Vault) AddPrivateKey(privateKey solana.PrivateKey) solana.PublicKey {
	v.KeyBag = append(v.KeyBag, privateKey)
	return privateKey.PublicKey()
}

// PrintPublicKeys prints a PublicKey corresponding to each PrivateKey in the Vault's
// KeyBag.
func (v *Vault) PrintPublicKeys() {
	fmt.Printf("Public keys contained within (%d in total):\n", len(v.KeyBag))
	for _, key := range v.KeyBag {
		fmt.Println("-", key.PublicKey().String())
	}
}

func (v *Vault) PrintPrivateKeys() {
	fmt.Printf("Private keys contained within (%d in total):\n", len(v.KeyBag))
	for _, key := range v.KeyBag {
		fmt.Printf("- %s (corresponds to %s)\n", key, key.PublicKey())
	}
}

// WriteToFile writes the Vault to disk. You need to encrypt before
// writing to file, otherwise you might lose much :)
func (v *Vault) WriteToFile(filename string) error {
	cnt, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fl, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	_, err = fl.Write(cnt)
	if err != nil {
		fl.Close()
		return err
	}

	return fl.Close()
}

func (v *Vault) Open(boxer SecretBoxer) error {
	data, err := boxer.Open(v.SecretBoxCiphertext)
	if err != nil {
		return fmt.Errorf("opening boxer: %w", err)
	}

	err = json.Unmarshal(data, &v.KeyBag)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}

func (v *Vault) Seal(boxer SecretBoxer) error {
	payload, err := json.Marshal(v.KeyBag)
	if err != nil {
		return err
	}

	v.SecretBoxWrap = boxer.WrapType()
	cipherText, err := boxer.Seal(payload)
	if err != nil {
		return err
	}

	v.SecretBoxCiphertext = cipherText
	return nil
}
