package zkencryption

import "github.com/solana-foundation/solana-go/v2"

// Signer is the minimal signing contract needed to derive confidential-transfer
// keys from a Solana signer. It is satisfied by solana.PrivateKey out of the
// box, and can be implemented by hardware wallets, remote signing services,
// or any other custody that can sign an arbitrary byte message and return a
// deterministic ed25519 signature.
type Signer interface {
	Sign(message []byte) (solana.Signature, error)
}
