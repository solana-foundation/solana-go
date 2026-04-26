# binary

A binary serialization package for Solana wire formats. Vendored from
`github.com/gagliardetto/binary` with substantial performance work on
top (see [BENCH.md](BENCH.md) for the headline numbers).

Handles three Solana-relevant encodings:

| Constant             | Used for                                                | Length prefix    |
| -------------------- | ------------------------------------------------------- | ---------------- |
| `EncodingBin`        | bincode-style (fluxd / legacy Solana tooling)           | uvarint          |
| `EncodingBorsh`      | Anchor programs, SPL state accounts, general Solana IDL | u32 LE           |
| `EncodingCompactU16` | Solana transaction / message length prefixes            | 1-3 byte compact |

Unless you know you need bincode, you almost always want `Borsh` for
program/account state and `CompactU16` for raw transaction/message
parsing.

---

## Quick start

Most callers only need the top-level marshal helpers. They work with
plain Go types and struct tags.

```go
import bin "github.com/gagliardetto/solana-go/binary"

type Foo struct {
    A uint64
    B string
    C []byte
}

// Encode.
wire, err := bin.MarshalBorsh(&Foo{A: 7, B: "hi", C: []byte{1,2,3}})

// Decode.
var out Foo
err = bin.UnmarshalBorsh(&out, wire)
```

The three encoding variants share the same signature:

```go
bin.MarshalBin(v any) ([]byte, error)        // uvarint lengths
bin.MarshalBorsh(v any) ([]byte, error)      // u32 LE lengths
bin.MarshalCompactU16(v any) ([]byte, error) // compact-u16 lengths

bin.UnmarshalBin(v any, b []byte) error
bin.UnmarshalBorsh(v any, b []byte) error
bin.UnmarshalCompactU16(v any, b []byte) error
```

These go through a reflection-driven encoder with a pooled internal
buffer. One allocation per call for the returned slice; zero per-field
allocations on the encode path.

---

## Struct tags

Fields control their wire layout through the `bin` struct tag. All tags
are space-separated tokens inside a single `bin:"..."` string.

| Tag                    | Effect                                                                                                                                         |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `sizeof=<fieldName>`   | This field's decoded value is the length of the named slice/array field later in the struct. Used when a separate count field precedes a body. |
| `big`                  | Encode/decode multi-byte integers as big-endian.                                                                                               |
| `little`               | Force little-endian (default already for borsh and compact-u16).                                                                               |
| `optional` or `option` | Field is preceded by a 1-byte "present" flag (Rust `Option<T>`).                                                                               |
| `coption`              | Field is preceded by a 4-byte "present" flag (Solana C-style Option).                                                                          |
| `binary_extension`     | Anchor "binary extension" sentinel: present-if-remaining-bytes.                                                                                |
| `skip` or `-`          | Skip this field on both encode and decode.                                                                                                     |
| `enum`                 | Struct is the tagged-union body of a borsh enum.                                                                                               |

Also recognized for compatibility with upstream borsh libs:

| Tag                 | Effect                  |
| ------------------- | ----------------------- |
| `borsh_skip:"true"` | Alias for `bin:"skip"`. |
| `borsh_enum:"true"` | Alias for `bin:"enum"`. |

### Example

```go
type Instruction struct {
    Discriminator uint8
    NumAccounts   uint8  `bin:"sizeof=AccountIdx"`
    AccountIdx    []uint8
    Data          []byte  // length-prefixed as bincode: uvarint
    Tip           *uint64 `bin:"optional"`
}
```

---

## Custom types (preferred for hot paths)

Types that can serialize themselves can implement the marshaler
interfaces. This is the mechanism `solana.PublicKey`, `solana.Signature`,
and most program types use internally.

```go
type BinaryMarshaler interface {
    MarshalWithEncoder(encoder *Encoder) error
}

type BinaryUnmarshaler interface {
    UnmarshalWithDecoder(decoder *Decoder) error
}
```

When both are implemented, the reflection encoder/decoder detects the
type and dispatches to the custom method -- no per-field reflection cost.
This is the fastest generic path and what program packages should
implement.

```go
func (p *Pubkey) MarshalWithEncoder(e *bin.Encoder) error {
    return e.WriteBytes(p[:], false)
}

func (p *Pubkey) UnmarshalWithDecoder(d *bin.Decoder) error {
    _, err := d.Read(p[:])
    return err
}
```

---

## Faster paths

The package exposes three zero-allocation escape hatches for latency-
sensitive code. Each one is progressively more unsafe in exchange for
more speed. Use the lowest-cost option that fits your constraints.

### 1. Pre-sized buffer encode -- `MarshalXxxInto`

If the caller already knows the wire size, the encoder can write
straight into the caller's buffer. Zero allocations, roughly 20-30%
faster than `Marshal*`.

```go
size, _ := bin.BorshByteCount(&foo)  // or compute statically
buf := make([]byte, size)

n, err := bin.MarshalBorshInto(&foo, buf)
wire := buf[:n]

// io.ErrShortBuffer is returned if buf is too small. buf is never
// reallocated -- this is the zero-alloc guarantee.
```

For repeat encodes into the same destination, keep an `*Encoder` and
re-target it with `ResetInto`:

```go
enc := bin.NewBorshEncoderInto(nil)
for _, msg := range messages {
    enc.ResetInto(scratch)
    _ = enc.Encode(msg)
    send(enc.Bytes())
}
```

### 2. Hand-rolled encoders -- `Cursor`

When you write the encoding logic yourself (e.g. hot-path program
instructions), `Cursor` skips the error-returning Encoder primitives
and does one memory poke per call. 6-12x faster than the Encoder for
primitive-heavy code; matches hand-written `binary.LittleEndian.Put*`
calls in generated assembly.

```go
buf := make([]byte, txHeaderSize)
wire := bin.NewCursor(buf).
    WriteU8(hdr.NumReqSigs).
    WriteU8(hdr.NumROSigned).
    WriteU8(hdr.NumROUnsigned).
    WriteLenCompactU16(len(accounts)).
    // ...
    Written()
```

Methods return `*Cursor`, so writes chain. A `Cursor` owns nothing --
out-of-bounds writes panic (standard Go slice bounds panic). Pre-size
correctly, or use `Encoder.NewXxxEncoderInto` for error-returning
writes.

Available methods: `WriteU8/16/32/64LE/BE`, `WriteI*`, `WriteBool`,
`WriteF32/64LE/BE`, `WriteBytes`, `WriteZero`, `Skip`, `WriteUvarint`,
`WriteVarint`, `WriteLenBin`, `WriteLenBorsh`, `WriteLenCompactU16`.
Back-patching via `SetPos(n)` / `Pos()` / `Written()`.

### 3. In-place mutation -- `ViewAs`

When the task is "patch field X in a pre-built wire buffer and send
it," do not decode, mutate, and re-encode. Reinterpret the buffer as a
typed pointer and write through it.

```go
// Patch a 32-byte recent blockhash at a known offset.
v, err := bin.ViewAs[bin.Blockhash](wire[blockhashOffset:])
if err != nil { return err }
*v = newBlockhash

// Mutations are visible in the original wire[] -- no copy.
send(wire)
```

`ViewAs[T]` returns `*T` aliasing the byte slice. `ViewSliceAs[T](buf, n)`
returns a `[]T` alias.

**Constraints.** T must be a POD type with no Go-compiler-inserted
padding. Run `bin.AssertPOD[T]()` once at program start (e.g. from
`init()`) to catch layout violations before they silently mis-patch
the wire:

```go
func init() {
    bin.MustAssertPOD[Blockhash]()
}
```

Safe shapes: fixed-size byte arrays (`[32]byte`, `[64]byte`), structs
of same-size integer fields, homogeneous fixed-size arrays. Unsafe:
mixed-width packed structs like `{uint8; uint64}` where Go adds 7
bytes of padding but the wire format is tight. `AssertPOD` flags these.
For packed formats, use `Cursor` with byte-level offsets instead.

~730x faster than a decode-then-encode round trip on a single-field
patch.

### 4. Generic memcpy marshal -- `MarshalPOD[T]` / `UnmarshalPOD[T]`

For types that satisfy the same POD constraints as `ViewAs`, the whole
struct can be marshaled or unmarshaled as a single `memcpy` rather than
field-by-field through reflection. Typed via Go generics -- no reflection
walk, no `any` boxing, no interface dispatch. The compiler often inlines
small fixed sizes into register moves.

```go
var key bin.Pubkey
for i := range key {
    key[i] = byte(i)
}

// Marshal into a pre-sized buffer (zero-alloc, detached copy of *v).
dst := make([]byte, 32)
n, err := bin.MarshalPOD(&key, dst)

// Or alloc-and-return for ergonomics:
wire := bin.MarshalPODAlloc(&key)

// Unmarshal: copy wire bytes into *v (detached from src).
var decoded bin.Pubkey
err = bin.UnmarshalPOD(&decoded, wire)
```

Unlike `ViewAs` (which aliases the source buffer), `MarshalPOD` and
`UnmarshalPOD` produce *detached* copies -- mutating one side after the
call does not affect the other. Use `ViewAs` when you want live alias
access; use `MarshalPOD` / `UnmarshalPOD` when you want a clean separation
between wire bytes and your owned struct.

Same POD constraints as `ViewAs`: run `AssertPOD[T]()` once at program
start to verify. Unsafe for padded structs, heterogeneous packed wire
formats, or big-endian hosts.

Benchmarks:

| Operation                | Reflective path | MarshalPOD path |
| ------------------------ | --------------: | --------------: |
| Marshal 32-byte Pubkey   | 58 ns           | ~0.25 ns (inlined to register moves) |
| Marshal 64-byte struct   | 118 ns          | ~0.25 ns |
| Unmarshal 64-byte struct | 60 ns           | ~0.76 ns |

See [BENCH.md](BENCH.md) for the full table.

---

## Bounding untrusted input

When decoding data from the network (RPC, websocket subscriptions,
untrusted block data) a malicious length prefix can trigger arbitrarily
large allocations. The decoder has two opt-in caps:

```go
dec := bin.NewBorshDecoder(payload).
    SetMaxSliceLen(256).  // reject slice prefixes > 256 elements
    SetMaxMapLen(64)      // reject map prefixes > 64 entries

if err := dec.Decode(&v); err != nil {
    // errors.Is(err, bin.ErrSliceLenTooLarge) == true on cap violation
}
```

Default (no caller-set cap) preserves historical behavior. Internally
the decoder also enforces `len * minElementWireSize <= Remaining()` so
a malicious `[]Pubkey` prefix claiming 1000 entries is rejected when
only 100 wire bytes remain, even with no explicit cap set.

---

## Picking a path

```
                             unknown size?
                                  |
                       +----------+----------+
                       |                     |
                     yes                    no
                       |                     |
                 Marshal/Unmarshal     MarshalBinInto
                 (simplest, 1 alloc)   (zero alloc)
                       |
               know the exact field
               layout + want no error
               returns?
                       |
                 +-----+-----+
                 |           |
                no          yes
                 |           |
             (stay with   Cursor
              Marshal)    (6-12x faster,
                          panics on OOB)
```

Separate decision for mutation:

```
need to patch
bytes in place?
     |
     |---> whole struct fits a POD shape? -> ViewAs[T]
     |                                       (+ AssertPOD[T] in init)
     |
     |---> mixed-width packed wire?        -> Cursor at known offsets
     |                                       + SetPos() for back-patch
     |
     +---> decoding and re-encoding        -> Marshal/Unmarshal round trip
           is acceptable                     (~100-200 ns overhead)
```

---

## Thread safety

`Encoder`, `Decoder`, and `Cursor` are not safe for concurrent use.
The top-level `Marshal*` / `Unmarshal*` helpers are safe to call from
multiple goroutines because they acquire their own pooled
Encoder/Decoder for each call.

---

## Reference

- Package docs: `go doc github.com/gagliardetto/solana-go/binary`
- Benchmarks: [BENCH.md](BENCH.md)
- Upstream (before vendoring): [github.com/gagliardetto/binary](https://github.com/gagliardetto/binary)
