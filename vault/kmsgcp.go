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
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"
)

///
/// Boxer implementation.
///

type KMSGCPBoxer struct {
	keyPath string
}

func NewKMSGCPBoxer(keyPath string) *KMSGCPBoxer {
	return &KMSGCPBoxer{
		keyPath: keyPath,
	}
}

func (b *KMSGCPBoxer) Seal(in []byte) (string, error) {
	mgr, err := NewKMSGCPManager(b.keyPath)
	if err != nil {
		return "", fmt.Errorf("new kms gcp manager: %w", err)
	}

	encrypted, err := mgr.Encrypt(in)
	if err != nil {
		return "", fmt.Errorf("kms encryption: %w", err)
	}

	return base64.RawStdEncoding.EncodeToString(encrypted), nil

}

func (b *KMSGCPBoxer) Open(in string) ([]byte, error) {
	mgr, err := NewKMSGCPManager(b.keyPath)
	if err != nil {
		return []byte{}, fmt.Errorf("new kms gcp manager: %w", err)
	}
	data, err := base64.RawStdEncoding.DecodeString(in)
	if err != nil {
		return []byte{}, fmt.Errorf("base 64 decode, %w", err)
	}
	out, err := mgr.Decrypt(data)
	if err != nil {
		return []byte{}, fmt.Errorf("base 64 decode, %w", err)
	}
	return out, nil
}

func (b *KMSGCPBoxer) WrapType() string {
	return "kms-gcp"
}

const (
	saltLength         = 16
	nonceLength        = 24
	keyLength          = 32
	shamirSecretLength = 32
)

func deriveKey(passphrase string, salt []byte) [keyLength]byte {
	secretKeyBytes := argon2.IDKey([]byte(passphrase), salt, 4, 64*1024, 4, 32)
	var secretKey [keyLength]byte
	copy(secretKey[:], secretKeyBytes)
	return secretKey
}

///
/// Manager implementation
///

func NewKMSGCPManager(keyPath string) (*KMSGCPManager, error) {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, cloudkms.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	kmsService, err := cloudkms.New(client)
	if err != nil {
		return nil, err
	}

	manager := &KMSGCPManager{
		service: kmsService,
		keyPath: keyPath,
	}

	return manager, nil
}

type KMSGCPManager struct {
	dekCache        map[string][32]byte
	dekCacheLock    sync.Mutex
	localDEK        [32]byte
	localWrappedDEK string
	service         *cloudkms.Service
	keyPath         string
}

func (k *KMSGCPManager) setupEncryption() error {
	if k.dekCache != nil {
		return nil
	}

	_, err := io.ReadFull(rand.Reader, k.localDEK[:])
	if err != nil {
		return err
	}

	req := &cloudkms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString(k.localDEK[:]),
	}

	resp, err := k.service.Projects.Locations.KeyRings.CryptoKeys.Encrypt(k.keyPath, req).Do()
	if err != nil {
		return err
	}

	k.localWrappedDEK = resp.Ciphertext
	k.dekCache = map[string][32]byte{resp.Ciphertext: k.localDEK}

	return nil
}

func (k *KMSGCPManager) fetchPlainDEK(wrappedDEK string) (out [32]byte, err error) {
	k.dekCacheLock.Lock()
	defer k.dekCacheLock.Unlock()

	if cachedKey, found := k.dekCache[wrappedDEK]; found {
		return cachedKey, nil
	}

	req := &cloudkms.DecryptRequest{
		Ciphertext: wrappedDEK,
	}
	resp, err := k.service.Projects.Locations.KeyRings.CryptoKeys.Decrypt(k.keyPath, req).Do()
	if err != nil {
		return
	}

	plainKey, err := base64.StdEncoding.DecodeString(resp.Plaintext)
	if err != nil {
		return
	}

	copy(out[:], plainKey)

	if k.dekCache == nil {
		k.dekCache = map[string][32]byte{}
	}
	if k.localWrappedDEK == "" {
		k.localWrappedDEK = wrappedDEK
	}
	k.dekCache[wrappedDEK] = out

	return
}

type BlobV1 struct {
	Version       int      `bson:"version"`
	WrappedDEK    string   `bson:"wrapped_dek"`
	Nonce         [24]byte `bson:"nonce"`
	EncryptedData []byte   `bson:"data"`
}

func (k *KMSGCPManager) Encrypt(in []byte) ([]byte, error) {
	if err := k.setupEncryption(); err != nil {
		return nil, err
	}

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	var sealedMsg []byte
	sealedMsg = secretbox.Seal(sealedMsg, in, &nonce, &k.localDEK)

	blob := &BlobV1{
		Version:       1,
		WrappedDEK:    k.localWrappedDEK,
		Nonce:         nonce,
		EncryptedData: sealedMsg,
	}

	cereal, err := json.Marshal(blob)
	if err != nil {
		return nil, err
	}

	return cereal, nil
}

func (k *KMSGCPManager) Decrypt(in []byte) ([]byte, error) {
	var blob BlobV1
	err := json.Unmarshal(in, &blob)
	if err != nil {
		return nil, err
	}

	// No need to check `blob.Version` == 1, we did it already with
	// the `magicFound` comparison.

	plainDEK, err := k.fetchPlainDEK(blob.WrappedDEK)
	if err != nil {
		return nil, err
	}

	plainData, ok := secretbox.Open(nil, blob.EncryptedData, &blob.Nonce, &plainDEK)
	if !ok {
		return nil, fmt.Errorf("failed decrypting data, that's all we know")
	}

	return plainData, nil
}
