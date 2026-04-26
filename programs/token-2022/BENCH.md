# programs/token-2022/ decode benchmarks

Comparison of four decode paths for Token-2022 `Mint` and `Account`
base layouts (identical to SPL-Token — Token-2022 extensions live in
a TLV trailer that these paths deliberately skip):

| Path          | Description                                                                                                                                         |
| ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Traditional` | Current flow: `bin.NewBinDecoder(data).Decode(&mint)` dispatching to `UnmarshalWithDecoder`.                                                        |
| `Raw`         | `DecodeMintRaw(&raw, data)` / `DecodeAccountRaw(&raw, data)` — POD memcpy into a raw struct.                                                        |
| `View`        | `ViewMintRaw(data)` / `ViewAccountRaw(data)` — zero-copy pointer alias into the wire buffer (TLV trailer safely ignored).                           |
| `Fast`        | `DecodeMintFast(data)` / `DecodeAccountFast(data)` — POD memcpy + materialization into the existing pointered `Mint`/`Account` (drop-in).           |

## Setup

- Host:     Apple M4 Max, darwin/arm64
- Runner:   `go test -bench 'BenchmarkDecodeMint|BenchmarkDecodeAccount' -benchmem -benchtime=1s -count=3 -run ^$ ./programs/token-2022/`
- Payload shapes:
    - `AllSet`: every COption populated. Worst case for the traditional decoder.
    - `NoOptionals`: Mint with no authorities.
    - `Typical`: Account with no optional fields. **Most common on-chain shape — this is what Geyser subscribers see at volume.**

## Mint decode (82 bytes on wire, plus optional TLV trailer)

| Benchmark                                      | ns/op  | B/op | allocs/op | vs Traditional |
| ---------------------------------------------- | -----: | ---: | --------: | -------------: |
| Traditional_AllSet                             | 70     | 192  | 4         | 1.0x           |
| Raw_AllSet                                     | 2.3    | 0    | 0         | **31x**        |
| View_AllSet                                    | 0.96   | 0    | 0         | **73x**        |
| Fast_AllSet                                    | 20     | 64   | 2         | 3.5x           |
| Traditional_NoOptionals                        | 50     | 128  | 2         | 1.0x           |
| Raw_NoOptionals                                | 2.3    | 0    | 0         | **22x**        |
| Fast_NoOptionals                               | 4.5    | 0    | 0         | **11x**        |

## Account decode (165 bytes on wire, plus optional TLV trailer)

| Benchmark                                      | ns/op  | B/op | allocs/op | vs Traditional |
| ---------------------------------------------- | -----: | ---: | --------: | -------------: |
| Traditional_AllSet                             | 99     | 280  | 5         | 1.0x           |
| Raw_AllSet                                     | 4.6    | 0    | 0         | **22x**        |
| View_AllSet                                    | 0.98   | 0    | 0         | **101x**       |
| Fast_AllSet                                    | 37     | 72   | 3         | 2.7x           |
| Traditional_Typical                            | 69     | 208  | 2         | 1.0x           |
| Raw_Typical                                    | 4.6    | 0    | 0         | **15x**        |
| View_Typical                                   | 0.99   | 0    | 0         | **70x**        |
| Fast_Typical                                   | 16     | 0    | 0         | **4.3x**       |

## Takeaways

- Numbers track SPL-Token exactly — which is the point: Token-2022 base
  layouts are byte-identical to SPL-Token, and the POD path reinterprets
  the same 82/165-byte prefix. Trailing TLV extension bytes are safely
  ignored by `ViewAs` (returns a prefix alias) and `UnmarshalPOD` (copies
  only `sizeof(T)`).

- **View path is sub-nanosecond** even when the buffer carries a
  multi-kB TLV trailer — the length of the source slice does not matter.

- **Fast path decoded into the pointered `Account` with no optionals
  (the `Typical` shape) is zero-alloc** and ~4x faster than the
  traditional decoder. This is the most common shape on-chain and the
  usual hot path for subscribers that do care about the public `Account`
  API.

- Extension data (TransferFeeConfig, ConfidentialTransfer, etc.) is
  **not** parsed by these paths. Use the extension decoders separately
  after reading the base image. Most subscribers only need the base
  fields (Amount, Owner, Mint, State) on each update, so the TLV
  parsing cost is avoided entirely.

## Trading-bot impact

A Geyser subscriber processing 10,000 Token-2022 account updates per
second:

| Path          | CPU per second | Allocs per second |
| ------------- | -------------: | ----------------: |
| Traditional   | ~690 us        | 20,000            |
| Fast          | ~160 us        | 0                 |
| View          | ~10 us         | 0                 |

Same picture as SPL-Token: the `View` path saves ~680 us/sec of CPU
and eliminates 20,000 allocations/sec, with the additional benefit
that it never touches the TLV trailer.

## Reproducing

```sh
go test -bench 'BenchmarkDecodeMint|BenchmarkDecodeAccount' \
    -benchmem -benchtime=1s -count=6 -run '^$' \
    ./programs/token-2022/
```

## Usage guidance

| Use case                                       | Recommended path                     |
| ---------------------------------------------- | ------------------------------------ |
| Read-only access, short-lived                  | `ViewMintRaw` / `ViewAccountRaw`     |
| Own-a-copy, still want the raw layout          | `DecodeMintRaw` / `DecodeAccountRaw` |
| Own-a-copy, need the existing `Mint`/`Account` | `DecodeMintFast` / `DecodeAccountFast` |
| Need to parse extensions too                   | Use the fast path for the base image, then extension decoders on the TLV trailer. |
| Don't care about latency                       | Existing `bin.NewBinDecoder(...).Decode(&v)` |

Parity is covered by `TestDecodeMintFast_ParityWithTraditional_*` and
`TestDecodeAccountFast_ParityWithTraditional_*`. TLV pass-through is
covered by `TestDecodeAccountFast_IgnoresTrailingTLV`.
