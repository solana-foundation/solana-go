//! Shared helpers for the token-2022 confidential extension parity generators.

use {solana_instruction::Instruction, solana_pubkey::Pubkey, std::num::NonZeroI8};

pub fn hex(b: &[u8]) -> String {
    b.iter().map(|x| format!("{:02x}", x)).collect()
}

/// pattern fills pods with distinct, non-zero bytes; seed keeps two values of
/// the same type distinguishable.
pub fn pattern(seed: u8, n: usize) -> Vec<u8> {
    (0..n)
        .map(|i| ((i + seed as usize) % 251 + 1) as u8)
        .collect()
}

pub fn pod<T: bytemuck::Pod>(seed: u8) -> T {
    *bytemuck::from_bytes(&pattern(seed, core::mem::size_of::<T>()))
}

pub fn addr(n: u8) -> Pubkey {
    Pubkey::from([n; 32])
}

pub fn ix_json(ix: &Instruction) -> String {
    let accounts: Vec<String> = ix
        .accounts
        .iter()
        .map(|a| {
            format!(
                r#"{{"pubkey":"{}","is_signer":{},"is_writable":{}}}"#,
                hex(a.pubkey.as_ref()),
                a.is_signer,
                a.is_writable
            )
        })
        .collect();
    format!(
        r#"{{"program_id":"{}","data":"{}","accounts":[{}]}}"#,
        hex(ix.program_id.as_ref()),
        hex(&ix.data),
        accounts.join(",")
    )
}

pub fn emit(name: &str, instructions: Vec<Instruction>) -> String {
    let encoded: Vec<String> = instructions.iter().map(ix_json).collect();
    format!(
        r#"{{"name":"{}","instructions":[{}]}}"#,
        name,
        encoded.join(",")
    )
}

pub fn offset(n: i8) -> NonZeroI8 {
    NonZeroI8::new(n).unwrap()
}

/// print_vectors writes the document the go parity tests read.
pub fn print_vectors(program_id: &Pubkey, vectors: Vec<String>) {
    println!(
        r#"{{"program_id":"{}","builders":[{}]}}"#,
        hex(program_id.as_ref()),
        vectors.join(",")
    );
}
