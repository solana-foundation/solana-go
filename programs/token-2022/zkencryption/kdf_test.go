package zkencryption_test

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token-2022/zkencryption"
)

type fromSeedVec struct {
	Name             string `json:"name"`
	SeedHex          string `json:"seed_hex"`
	AeKeyHex         string `json:"ae_key_hex"`
	ElGamalSecretHex string `json:"elgamal_secret_hex"`
}

type fromSignatureVec struct {
	Name             string `json:"name"`
	SignatureHex     string `json:"signature_hex"`
	AeKeyHex         string `json:"ae_key_hex"`
	ElGamalSecretHex string `json:"elgamal_secret_hex"`
}

type fromSignerVec struct {
	Name             string `json:"name"`
	KeypairSecretHex string `json:"keypair_secret_hex"`
	PublicSeedHex    string `json:"public_seed_hex"`
	AeKeyHex         string `json:"ae_key_hex"`
	ElGamalSecretHex string `json:"elgamal_secret_hex"`
}

type fromMnemonicVec struct {
	Name             string `json:"name"`
	Mnemonic         string `json:"mnemonic"`
	Passphrase       string `json:"passphrase"`
	AeKeyHex         string `json:"ae_key_hex"`
	ElGamalSecretHex string `json:"elgamal_secret_hex"`
}

type vectors struct {
	FromSeed      []fromSeedVec      `json:"from_seed"`
	FromSignature []fromSignatureVec `json:"from_signature"`
	FromSigner    []fromSignerVec    `json:"from_signer"`
	FromMnemonic  []fromMnemonicVec  `json:"from_mnemonic"`
}

func loadVectors(t *testing.T) vectors {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "kdf_vectors.json"))
	if err != nil {
		t.Fatalf("read test vectors: %v", err)
	}
	var v vectors
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("decode test vectors: %v", err)
	}
	return v
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("decode hex %q: %v", s, err)
	}
	return b
}

func TestFromSeed_MatchesRust(t *testing.T) {
	t.Parallel()
	for _, v := range loadVectors(t).FromSeed {
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			seed := mustHex(t, v.SeedHex)

			ae, err := zkencryption.AeKeyFromSeed(seed)
			if err != nil {
				t.Fatalf("AeKeyFromSeed: %v", err)
			}
			if got := hex.EncodeToString(ae[:]); got != v.AeKeyHex {
				t.Errorf("AeKey mismatch: got %s want %s", got, v.AeKeyHex)
			}

			el, err := zkencryption.ElGamalSecretKeyFromSeed(seed)
			if err != nil {
				t.Fatalf("ElGamalSecretKeyFromSeed: %v", err)
			}
			if got := hex.EncodeToString(el[:]); got != v.ElGamalSecretHex {
				t.Errorf("ElGamalSecretKey mismatch: got %s want %s", got, v.ElGamalSecretHex)
			}
		})
	}
}

func TestFromSignature_MatchesRust(t *testing.T) {
	t.Parallel()
	for _, v := range loadVectors(t).FromSignature {
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			raw := mustHex(t, v.SignatureHex)
			var sig solana.Signature
			copy(sig[:], raw)

			ae, err := zkencryption.AeKeyFromSignature(sig)
			if err != nil {
				t.Fatalf("AeKeyFromSignature: %v", err)
			}
			if got := hex.EncodeToString(ae[:]); got != v.AeKeyHex {
				t.Errorf("AeKey mismatch: got %s want %s", got, v.AeKeyHex)
			}

			el, err := zkencryption.ElGamalSecretKeyFromSignature(sig)
			if err != nil {
				t.Fatalf("ElGamalSecretKeyFromSignature: %v", err)
			}
			if got := hex.EncodeToString(el[:]); got != v.ElGamalSecretHex {
				t.Errorf("ElGamalSecretKey mismatch: got %s want %s", got, v.ElGamalSecretHex)
			}
		})
	}
}

func TestFromSigner_MatchesRust(t *testing.T) {
	t.Parallel()
	for _, v := range loadVectors(t).FromSigner {
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			seed32 := mustHex(t, v.KeypairSecretHex)
			if len(seed32) != ed25519.SeedSize {
				t.Fatalf("keypair seed is %d bytes, want %d", len(seed32), ed25519.SeedSize)
			}
			// solana.PrivateKey matches ed25519.PrivateKey layout: seed || pubkey.
			priv := solana.PrivateKey(ed25519.NewKeyFromSeed(seed32))
			publicSeed := mustHex(t, v.PublicSeedHex)

			ae, err := zkencryption.AeKeyFromSigner(priv, publicSeed)
			if err != nil {
				t.Fatalf("AeKeyFromSigner: %v", err)
			}
			if got := hex.EncodeToString(ae[:]); got != v.AeKeyHex {
				t.Errorf("AeKey mismatch: got %s want %s", got, v.AeKeyHex)
			}

			el, err := zkencryption.ElGamalSecretKeyFromSigner(priv, publicSeed)
			if err != nil {
				t.Fatalf("ElGamalSecretKeyFromSigner: %v", err)
			}
			if got := hex.EncodeToString(el[:]); got != v.ElGamalSecretHex {
				t.Errorf("ElGamalSecretKey mismatch: got %s want %s", got, v.ElGamalSecretHex)
			}
		})
	}
}

func TestFromSeedPhraseAndPassphrase_MatchesRust(t *testing.T) {
	t.Parallel()
	for _, v := range loadVectors(t).FromMnemonic {
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			ae, err := zkencryption.AeKeyFromSeedPhraseAndPassphrase(v.Mnemonic, v.Passphrase)
			if err != nil {
				t.Fatalf("AeKeyFromSeedPhraseAndPassphrase: %v", err)
			}
			if got := hex.EncodeToString(ae[:]); got != v.AeKeyHex {
				t.Errorf("AeKey mismatch: got %s want %s", got, v.AeKeyHex)
			}

			el, err := zkencryption.ElGamalSecretKeyFromSeedPhraseAndPassphrase(v.Mnemonic, v.Passphrase)
			if err != nil {
				t.Fatalf("ElGamalSecretKeyFromSeedPhraseAndPassphrase: %v", err)
			}
			if got := hex.EncodeToString(el[:]); got != v.ElGamalSecretHex {
				t.Errorf("ElGamalSecretKey mismatch: got %s want %s", got, v.ElGamalSecretHex)
			}
		})
	}
}

func TestSeedLengthBounds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		run     func() error
		wantErr error
	}{
		{
			"ae_too_short",
			func() error { _, err := zkencryption.AeKeyFromSeed(make([]byte, 15)); return err },
			zkencryption.ErrSeedTooShort,
		},
		{
			"ae_at_minimum",
			func() error { _, err := zkencryption.AeKeyFromSeed(make([]byte, 16)); return err },
			nil,
		},
		{
			"ae_too_long",
			func() error { _, err := zkencryption.AeKeyFromSeed(make([]byte, 65536)); return err },
			zkencryption.ErrSeedTooLong,
		},
		{
			"elgamal_too_short",
			func() error { _, err := zkencryption.ElGamalSecretKeyFromSeed(make([]byte, 31)); return err },
			zkencryption.ErrSeedTooShort,
		},
		{
			"elgamal_at_minimum",
			func() error { _, err := zkencryption.ElGamalSecretKeyFromSeed(make([]byte, 32)); return err },
			nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.run()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// stubSigner implements zkencryption.Signer with a fixed signature, for
// exercising error paths (notably the default-signature rejection) without
// needing a specially crafted ed25519 keypair.
type stubSigner struct {
	sig solana.Signature
	err error
}

func (s stubSigner) Sign(_ []byte) (solana.Signature, error) { return s.sig, s.err }

func TestFromSigner_RejectsDefaultSignature(t *testing.T) {
	t.Parallel()
	signer := stubSigner{} // zero-valued Signature

	if _, err := zkencryption.AeKeyFromSigner(signer, []byte("seed")); !errors.Is(err, zkencryption.ErrDefaultSignature) {
		t.Fatalf("AeKeyFromSigner: err = %v, want ErrDefaultSignature", err)
	}
	if _, err := zkencryption.ElGamalSecretKeyFromSigner(signer, []byte("seed")); !errors.Is(err, zkencryption.ErrDefaultSignature) {
		t.Fatalf("ElGamalSecretKeyFromSigner: err = %v, want ErrDefaultSignature", err)
	}
}

func TestFromSigner_WrapsSignerError(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("hsm unreachable")
	signer := stubSigner{err: sentinel}

	if _, err := zkencryption.AeKeyFromSigner(signer, nil); !errors.Is(err, sentinel) {
		t.Fatalf("AeKeyFromSigner: err = %v, want wrapped sentinel", err)
	}
	if _, err := zkencryption.ElGamalSecretKeyFromSigner(signer, nil); !errors.Is(err, sentinel) {
		t.Fatalf("ElGamalSecretKeyFromSigner: err = %v, want wrapped sentinel", err)
	}
}
