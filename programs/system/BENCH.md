# programs/system/ decode benchmarks

Comparison of three decode paths for the system program's durable-nonce
account (`NonceAccount`, 80 bytes on wire).

| Path          | Description                                                                             |
| ------------- | --------------------------------------------------------------------------------------- |
| `Traditional` | `bin.NewBinDecoder(data).Decode(&n)` dispatching through `UnmarshalWithDecoder`.        |
| `Fast`        | `DecodeNonceAccountFast(data)` — single `bin.UnmarshalPOD` memcpy into a detached copy. |
| `View`        | `ViewNonceAccount(data)` — zero-copy pointer alias into the wire buffer.                |

## Setup

- Host:     Apple M4 Max, darwin/arm64
- Runner:   `go test -bench BenchmarkDecodeNonceAccount -benchmem -benchtime=1s -count=3 -run ^$ ./programs/system/`
- Layout:   `u32 Version | u32 State | [32]byte AuthorizedPubkey | [32]byte Nonce | u64 LamportsPerSignature` (80 bytes total, no COptions).

## NonceAccount decode

| Benchmark                          | ns/op | B/op | allocs/op | vs Traditional |
| ---------------------------------- | ----: | ---: | --------: | -------------: |
| DecodeNonceAccount_Traditional     | 52    | 176  | 2         | 1.0x           |
| DecodeNonceAccount_Fast            | 2.2   | 0    | 0         | **24x**        |
| DecodeNonceAccount_View            | 0.99  | 0    | 0         | **52x**        |

## Takeaways

- NonceAccount has no COption fields, so the payoff is pure decode
  throughput and allocation elimination (the 176 B / 2 allocs per call
  on the traditional path come from the decoder's internal scratch,
  not the struct itself).

- **Fast path is a 24x speedup** with zero allocations and is a
  drop-in: it returns the same `NonceAccount` value the traditional
  decoder produces.

- **View path** lets a caller read-through the wire buffer without
  copying the 80 bytes at all. Useful when only one or two fields are
  needed per update (e.g. just the `Nonce` pubkey).

- The layout is simple enough that the existing `NonceAccount` struct
  is itself POD-safe: `init()` runs `bin.MustAssertPOD[NonceAccount]()`,
  so a future field reorder that would introduce Go padding is caught
  immediately.

## Reproducing

```sh
go test -bench BenchmarkDecodeNonceAccount \
    -benchmem -benchtime=1s -count=6 -run '^$' \
    ./programs/system/
```

## Usage guidance

| Use case                                                    | Recommended path                 |
| ----------------------------------------------------------- | -------------------------------- |
| Read-only access, short-lived (watcher reads `Nonce` only)  | `ViewNonceAccount`               |
| Need a detached copy of the whole struct                    | `DecodeNonceAccountFast`         |
| Writing into a caller-owned destination (hot loop)          | `DecodeNonceAccountRaw(&n, buf)` |
| Don't care about latency                                    | Existing `bin.NewBinDecoder(...).Decode(&n)` |

Parity is covered by `TestDecodeNonceAccountFast_ParityWithTraditional`
and `TestMarshalPOD_RoundTripNonceAccount`.
