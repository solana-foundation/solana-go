# programs/token/ decode benchmarks

Comparison of three decode paths for SPL-Token `Mint` and `Account`
layouts:

| Path          | Description                                                                                   |
| ------------- | --------------------------------------------------------------------------------------------- |
| `Traditional` | Current flow: `bin.NewBinDecoder(data).Decode(&mint)` dispatching to `UnmarshalWithDecoder`.  |
| `Raw`         | `DecodeMintRaw(&raw, data)` / `DecodeAccountRaw(&raw, data)` — POD memcpy into a raw struct.  |
| `View`        | `ViewMintRaw(data)` / `ViewAccountRaw(data)` — zero-copy pointer alias into the wire buffer.  |
| `Fast`        | `DecodeMintFast(data)` / `DecodeAccountFast(data)` — POD memcpy + materialization into the existing pointered `Mint`/`Account` (drop-in replacement for `Traditional`). |

## Setup

- Host:     Apple M4 Max, darwin/arm64
- Runner:   `go test -bench 'BenchmarkDecodeMint|BenchmarkDecodeAccount' -benchmem -benchtime=1s -count=3 -run ^$ ./programs/token/`
- Payload shapes:
    - `AllSet`: every COption populated (2 authorities on Mint; delegate + is_native + close_authority on Account). Worst case for the traditional decoder.
    - `NoOptionals`: Mint with no authorities.
    - `Typical`: Account with no optional fields. **Most common on-chain shape — this is what Geyser subscribers see at volume.**

## Mint decode (82 bytes on wire)

| Benchmark                                      | ns/op  | B/op | allocs/op | vs Traditional |
| ---------------------------------------------- | -----: | ---: | --------: | -------------: |
| Traditional_AllSet                             | 71     | 192  | 4         | 1.0x           |
| Raw_AllSet                                     | 2.3    | 0    | 0         | **31x**        |
| View_AllSet                                    | 0.99   | 0    | 0         | **72x**        |
| Fast_AllSet                                    | 19     | 64   | 2         | 3.7x           |
| Traditional_NoOptionals                        | 51     | 128  | 2         | 1.0x           |
| Raw_NoOptionals                                | 2.3    | 0    | 0         | **22x**        |
| Fast_NoOptionals                               | 4.6    | 0    | 0         | **11x**        |

## Account decode (165 bytes on wire)

| Benchmark                                      | ns/op  | B/op | allocs/op | vs Traditional |
| ---------------------------------------------- | -----: | ---: | --------: | -------------: |
| Traditional_AllSet                             | 96     | 280  | 5         | 1.0x           |
| Raw_AllSet                                     | 4.6    | 0    | 0         | **21x**        |
| View_AllSet                                    | 0.99   | 0    | 0         | **97x**        |
| Fast_AllSet                                    | 36     | 72   | 3         | 2.7x           |
| Traditional_Typical                            | 67     | 208  | 2         | 1.0x           |
| Raw_Typical                                    | 4.6    | 0    | 0         | **15x**        |
| View_Typical                                   | 0.99   | 0    | 0         | **68x**        |
| Fast_Typical                                   | 16     | 0    | 0         | **4.1x**       |

## Takeaways

- **View path is sub-nanosecond** on both AllSet and Typical, with zero
  allocations. The whole struct is a single pointer cast; every accessor
  inlines to one unaligned load.

- **Raw decode is ~2-5 ns**, bounded by the `memcpy` of 82/165 bytes.
  Zero allocations. Detached copy (safe to mutate the source buffer
  afterward).

- **Fast decode (drop-in to `Mint`/`Account`)** pays for the
  `*PublicKey` allocations required by the existing pointered struct —
  each present COption costs one heap allocation. When all optionals
  are *absent* (the `Typical` Account shape), there are no allocations
  at all and it's 4x faster than the traditional decoder while remaining
  a drop-in.

- **Traditional decode** allocates 128-280 bytes per call and runs
  50-96 ns. These numbers double-count a bit because each `*PublicKey`
  is separately allocated; the bounds checks across the hand-written
  decoder add further overhead.

## Trading-bot impact

A Geyser subscriber processing 10,000 account updates per second
(typical for a Solana market-making or arbitrage bot) spends:

| Path          | CPU per second | Allocs per second |
| ------------- | -------------: | ----------------: |
| Traditional   | ~670 us        | 20,000            |
| Fast          | ~160 us        | 0                 |
| View          | ~10 us         | 0                 |

The `View` path saves ~660 us/sec of CPU and eliminates 20,000
allocations/sec — which translates to less GC pressure, lower tail
latency on the hot decision path, and roughly an order of magnitude
more headroom before the subscriber becomes the bottleneck.

## Reproducing

```sh
go test -bench 'BenchmarkDecodeMint|BenchmarkDecodeAccount' \
    -benchmem -benchtime=1s -count=6 -run '^$' \
    ./programs/token/

go install golang.org/x/perf/cmd/benchstat@latest
~/go/bin/benchstat <output>
```

## Usage guidance

| Use case                                       | Recommended path                     |
| ---------------------------------------------- | ------------------------------------ |
| Read-only access, short-lived                  | `ViewMintRaw` / `ViewAccountRaw`     |
| Own-a-copy, still want the raw layout          | `DecodeMintRaw` / `DecodeAccountRaw` |
| Own-a-copy, need the existing `Mint`/`Account` | `DecodeMintFast` / `DecodeAccountFast` |
| Don't care about latency                       | Existing `bin.NewBinDecoder(...).Decode(&v)` |

All four paths return equivalent data (for matching optional fields,
`View` and `Raw` expose it through accessor methods rather than
pointers). Parity is covered by `TestDecodeMintFast_ParityWithTraditional_*`
and `TestDecodeAccountFast_ParityWithTraditional_*`.
