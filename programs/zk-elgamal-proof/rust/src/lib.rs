//! wasm32 interface over `solana-zk-sdk` and `solana-zk-elgamal-proof-interface` zk-proof functionality.
mod aes;
mod constants;
mod el_gamal_encrypt;
mod grp_elgamal_encrypt;
mod memory;
mod parsing;
mod pedersen;
mod proof_gen;
mod verify;
