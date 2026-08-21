use std::alloc::{alloc, dealloc, Layout};

use bytemuck::Pod;

use crate::constants::ERR_OOM;

// ABI assumes wasm 32-bit addressing
#[cfg(not(target_pointer_width = "32"))]
compile_error!("solana-zk-sdk-wasm assumes 32-bit pointers");

const MARKER: u32 = 0x5A4B_0001; // 'ZK' | 0001
const HEADER_LEN: usize = 8;
const ALIGN: usize = 8;

#[no_mangle]
pub extern "C" fn zk_alloc(len: u32) -> u32 {
    raw_alloc(len)
}

/// Allocate `len` usable bytes, returning a pointer past the header (0 on failure).
fn raw_alloc(len: u32) -> u32 {
    let total = match HEADER_LEN.checked_add(len as usize) {
        Some(total) => total,
        None => return 0,
    };
    unsafe {
        let layout = Layout::from_size_align_unchecked(total, ALIGN);
        let base = alloc(layout);
        if base.is_null() {
            return 0;
        }
        (base as *mut u32).write(MARKER);
        (base.add(4) as *mut u32).write(len);
        base.add(HEADER_LEN) as u32
    }
}

/// Free a buffer returned by `zk_alloc` or `stash`.
#[no_mangle]
pub unsafe extern "C" fn zk_free(ptr: u32) {
    let Some(len) = buffer_len(ptr) else {
        return;
    };
    let base = (ptr as usize - HEADER_LEN) as *mut u8;
    (base as *mut u32).write(0); // Make sure double free fails the marker check.
    let layout = Layout::from_size_align_unchecked(HEADER_LEN + len as usize, ALIGN);
    dealloc(base, layout);
}

/// Return the length recorded in a buffer's header, or `None` if invalid.
pub(crate) unsafe fn buffer_len(ptr: u32) -> Option<u32> {
    if ptr == 0 || (ptr as usize) < HEADER_LEN {
        return None;
    }
    let base = (ptr as usize - HEADER_LEN) as *const u8;
    if (base as *const u32).read() != MARKER {
        return None;
    }
    Some((base.add(4) as *const u32).read())
}

/// Copy `bytes` into a result buffer and return a packed descriptor `(len << 32) | ptr`.
///
/// Returns negative `ERR_*` code on failure.
pub(crate) fn stash(bytes: &[u8]) -> i64 {
    if bytes.len() > i32::MAX as usize {
        return ERR_OOM as i64;
    }
    let len = bytes.len() as u32;
    let ptr = raw_alloc(len);
    if ptr == 0 {
        return ERR_OOM as i64;
    }
    unsafe {
        std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr as *mut u8, len as usize);
    }
    ((len as i64) << 32) | ptr as i64
}

pub(crate) fn stash_u64(value: u64) -> i64 {
    stash(&value.to_le_bytes())
}

pub(crate) fn stash_res<T: Pod, E, const ERR_CODE: i32>(res: Result<T, E>) -> i64 {
    match res {
        Ok(data) => stash_pod(&data),
        Err(_) => ERR_CODE.into(),
    }
}

pub(crate) fn stash_pod<T: Pod>(value: &T) -> i64 {
    stash(bytemuck::bytes_of(value))
}
