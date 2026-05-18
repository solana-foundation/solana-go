# Cross-language KDF test vectors

`kdf_vectors.json` holds the authoritative byte-for-byte outputs of the
Rust `solana-zk-sdk` key-derivation functions for a fixed set of inputs.
The Go tests in the parent package load this file and assert equality,
which is how we guarantee our Go KDF stays aligned with the Rust / JS /
WASM SDKs.

The fixture is static. Regeneration is only needed if `solana-zk-sdk`
ships a breaking change to the KDF (historically stable).

## What the vectors cover

- `from_seed`: `AeKey::from_seed` / `ElGamalSecretKey::from_seed` on a set
  of fixed-length entropy seeds (32, 48, 64 bytes).
- `from_signature`: `new_from_signature` on deterministic signature byte
  patterns (no signer involvement).
- `from_signer`: `new_from_signer` using fixed ed25519 keypairs and a
  range of public-seed values (empty, short, long).
- `from_mnemonic`: `from_seed_phrase_and_passphrase` on the official
  Trezor/BIP39 mnemonics with and without a passphrase.

## Regenerating

Requires a working Rust toolchain. Create a throwaway cargo project and
run the harness below; replace the existing `kdf_vectors.json` with its
stdout.

`Cargo.toml`:

```toml
[package]
name = "kdf-vector-gen"
version = "0.0.0"
edition = "2021"
publish = false

[dependencies]
solana-zk-sdk = "6.0.1"
solana-signature = "3.1.0"
solana-signer = "3.0.0"
solana-keypair = "3.0.1"
solana-seed-derivable = "3.0.0"
serde = { version = "1", features = ["derive"] }
serde_json = "1"
hex = "0.4"
```

`src/main.rs`:

```rust
use serde::Serialize;
use solana_keypair::Keypair;
use solana_seed_derivable::SeedDerivable;
use solana_signature::Signature;
use solana_zk_sdk::encryption::auth_encryption::AeKey;
use solana_zk_sdk::encryption::elgamal::ElGamalSecretKey;

#[derive(Serialize)]
struct FromSeedVector {
    name: String,
    seed_hex: String,
    ae_key_hex: String,
    elgamal_secret_hex: String,
}

#[derive(Serialize)]
struct FromSignatureVector {
    name: String,
    signature_hex: String,
    ae_key_hex: String,
    elgamal_secret_hex: String,
}

#[derive(Serialize)]
struct FromSignerVector {
    name: String,
    keypair_secret_hex: String,
    public_seed_hex: String,
    ae_key_hex: String,
    elgamal_secret_hex: String,
}

#[derive(Serialize)]
struct FromMnemonicVector {
    name: String,
    mnemonic: String,
    passphrase: String,
    ae_key_hex: String,
    elgamal_secret_hex: String,
}

#[derive(Serialize)]
struct Vectors {
    from_seed: Vec<FromSeedVector>,
    from_signature: Vec<FromSignatureVector>,
    from_signer: Vec<FromSignerVector>,
    from_mnemonic: Vec<FromMnemonicVector>,
}

fn ae_bytes(key: &AeKey) -> [u8; 16] {
    key.clone().into()
}

fn derive_from_seed(name: &str, seed: &[u8]) -> FromSeedVector {
    let ae = AeKey::from_seed(seed).unwrap();
    let el = ElGamalSecretKey::from_seed(seed).unwrap();
    FromSeedVector {
        name: name.into(),
        seed_hex: hex::encode(seed),
        ae_key_hex: hex::encode(ae_bytes(&ae)),
        elgamal_secret_hex: hex::encode(el.as_bytes()),
    }
}

fn derive_from_signature(name: &str, bytes: [u8; 64]) -> FromSignatureVector {
    let sig = Signature::from(bytes);
    let ae = AeKey::new_from_signature(&sig).unwrap();
    let el = ElGamalSecretKey::new_from_signature(&sig).unwrap();
    FromSignatureVector {
        name: name.into(),
        signature_hex: hex::encode(bytes),
        ae_key_hex: hex::encode(ae_bytes(&ae)),
        elgamal_secret_hex: hex::encode(el.as_bytes()),
    }
}

fn derive_from_signer(name: &str, seed32: &[u8; 32], public_seed: &[u8]) -> FromSignerVector {
    let kp = Keypair::new_from_array(*seed32);
    let ae = AeKey::new_from_signer(&kp, public_seed).unwrap();
    let el = ElGamalSecretKey::new_from_signer(&kp, public_seed).unwrap();
    FromSignerVector {
        name: name.into(),
        keypair_secret_hex: hex::encode(seed32),
        public_seed_hex: hex::encode(public_seed),
        ae_key_hex: hex::encode(ae_bytes(&ae)),
        elgamal_secret_hex: hex::encode(el.as_bytes()),
    }
}

fn derive_from_mnemonic(name: &str, mnemonic: &str, passphrase: &str) -> FromMnemonicVector {
    let ae = AeKey::from_seed_phrase_and_passphrase(mnemonic, passphrase).unwrap();
    let el = ElGamalSecretKey::from_seed_phrase_and_passphrase(mnemonic, passphrase).unwrap();
    FromMnemonicVector {
        name: name.into(),
        mnemonic: mnemonic.into(),
        passphrase: passphrase.into(),
        ae_key_hex: hex::encode(ae_bytes(&ae)),
        elgamal_secret_hex: hex::encode(el.as_bytes()),
    }
}

fn main() {
    let from_seed = vec![
        derive_from_seed("all_zero_32", &[0u8; 32]),
        derive_from_seed("all_one_32", &[0x11u8; 32]),
        derive_from_seed("ascending_64", &(0u8..64).collect::<Vec<u8>>()),
        derive_from_seed(
            "mixed_48",
            &hex::decode(
                "0102030405060708090a0b0c0d0e0f101112131415161718\
                 191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f30",
            )
            .unwrap(),
        ),
    ];

    let from_signature = vec![
        derive_from_signature("sig_deadbeef", {
            let mut s = [0u8; 64];
            for (i, b) in s.iter_mut().enumerate() {
                *b = ((i as u16 * 7 + 0xde) & 0xff) as u8;
            }
            s
        }),
        derive_from_signature("sig_increment", {
            let mut s = [0u8; 64];
            for (i, b) in s.iter_mut().enumerate() {
                *b = i as u8;
            }
            s
        }),
    ];

    let kp_seed_a: [u8; 32] = [
        0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
        0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee,
        0xff, 0x00,
    ];
    let kp_seed_b: [u8; 32] = [
        0x7e, 0x57, 0x7e, 0x57, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44,
        0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33,
        0x44, 0x55,
    ];

    let from_signer = vec![
        derive_from_signer("keypair_a_empty_seed", &kp_seed_a, &[]),
        derive_from_signer("keypair_a_mint_seed", &kp_seed_a, b"some-mint-pubkey"),
        derive_from_signer("keypair_b_mint_seed", &kp_seed_b, b"some-mint-pubkey"),
        derive_from_signer(
            "keypair_b_long_seed",
            &kp_seed_b,
            b"a-rather-longer-public-seed-used-for-domain-separation",
        ),
    ];

    let from_mnemonic = vec![
        derive_from_mnemonic(
            "trezor_all_abandon_no_passphrase",
            "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
            "",
        ),
        derive_from_mnemonic(
            "trezor_all_abandon_trezor_passphrase",
            "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
            "TREZOR",
        ),
        derive_from_mnemonic(
            "legal_winner_trezor",
            "legal winner thank year wave sausage worth useful legal winner thank yellow",
            "TREZOR",
        ),
    ];

    let out = Vectors { from_seed, from_signature, from_signer, from_mnemonic };
    println!("{}", serde_json::to_string_pretty(&out).unwrap());
}
```

Then:

```
cargo run --release > path/to/kdf_vectors.json
```
