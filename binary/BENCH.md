# binary/ performance baseline

Comparison of the vendored `binary` package at the first branch commit
(before any perf work on top of it) against HEAD (all branch perf work,
including techniques #2, #4, #7, #8 added on top of the initial vendoring
and the in-branch perf refactors).

## Setup

- Initial: `56bb04765b227a498a22a9a7f47a4c35a11c7576` ("perf: vendor and improve binary pkg")
- HEAD:    `d736ed98a0789c29ca6fc46ba5b010c86a351c80`
- Host:    Apple M4 Max, darwin/arm64
- Runner:  `go test -bench . -benchmem -benchtime=500ms -count=6 -run ^$ ./binary/`
- Stats:   `benchstat` (6 runs per benchmark)

## Headline wins (shared benchmarks, present on both sides)

| Benchmark                         | Initial   | HEAD      | Time     | B/op              | allocs/op       |
| --------------------------------- | --------: | --------: | -------: | ----------------: | --------------: |
| Encode_Struct_Borsh               | 248.8 ns  | 199.2 ns  | -19.9%   | 248 -> 112 (-55%) | 4 -> 1 (-75%)   |
| Encode_Struct_Borsh_Buffered      | 240.3 ns  | 193.1 ns  | -19.7%   | 136 -> 0 (-100%)  | 3 -> 0 (-100%)  |
| ByteCount/flat                    | 226.1 ns  | 156.4 ns  | -30.8%   | 216 -> 120 (-44%) | 6 -> 2 (-67%)   |
| ByteCount/nested/small_list       | 1385 ns   | 919 ns    | -33.7%   | 720 -> 184 (-74%) | 41 -> 10 (-76%) |
| ByteCount/nested/large_list       | 17.48 us  | 12.85 us  | -26.5%   | -31%              | -67%            |
| ByteCount/deep/small_list         | 4.42 us   | 2.79 us   | -36.7%   | 2048 -> 312 (-85%) | 123 -> 26 (-79%) |
| ByteCount/deep/large_list         | 52.92 us  | 38.69 us  | -26.9%   | -31%              | -67%            |
| CompactU16 (reader)               | 1.26 ns   | 1.23 ns   | -2.3%    | -                 | -               |
| CompactU16Encode                  | 10.09 ns  | 9.47 ns   | -6.2%    | -                 | -               |
| _uintSlice32_Decode_field_withCustomDecoder | 2.52 us | 2.47 us | -1.9% | -            | -               |

## Small regressions (micro-bench primitives)

Sub-nanosecond absolute regressions on single-primitive writes. Root
cause is the `if e.fixedBuf && ...` branch added in `toWriter` for the
fixed-buffer mode (#2). The branch predicts perfectly when fixed mode
isn't in use, but it still requires one byte load; at the 3 ns granularity
of a single WriteUintN call this shows up as +0.3-0.6 ns.

For hot loops that accumulate this cost, `Cursor` (#4) is the escape
valve: it skips the Encoder primitives entirely and is ~12x faster for
primitive-heavy workloads.

| Benchmark                         | Initial   | HEAD      | Delta        |
| --------------------------------- | --------: | --------: | -----------: |
| Encode_WriteUint16                | 3.09 ns   | 3.57 ns   | +15.4% (+0.5 ns) |
| Encode_WriteUint32                | 3.09 ns   | 3.61 ns   | +16.9% (+0.5 ns) |
| Encode_WriteUint64                | 3.06 ns   | 3.67 ns   | +19.8% (+0.6 ns) |
| Encode_WriteUint64_Buffered       | 3.85 ns   | 4.24 ns   | +10.3% (+0.4 ns) |
| Encode_CompactU16_1byte           | 6.51 ns   | 6.85 ns   | +5.2%  (+0.3 ns) |
| Encode_CompactU16_2byte           | 6.45 ns   | 7.02 ns   | +8.8%  (+0.6 ns) |
| Decode_SliceUint64_8k             | 4.27 us   | 4.48 us   | +4.8%        |
| Decode_SliceUint32_8k             | 2.36 us   | 2.44 us   | +3.2%        |
| Decode_ReadString_Copy            | 29.7 ns   | 31.2 ns   | +5.3%        |
| Decode_ReadString_Borrow          | 19.96 ns  | 21.11 ns  | +5.8%        |

## HEAD-only (new capabilities)

Benchmarks for APIs introduced by techniques #2, #4, #7. No baseline
exists on the initial commit. Reported for reference and as the reason
the small primitive regressions are acceptable.

| Benchmark                         | ns/op     | B/op  | allocs/op | Technique |
| --------------------------------- | --------: | ----: | --------: | --------- |
| MarshalInto_Struct_Borsh          | 200.4     | 0     | 0         | #2 EncodeInto |
| Marshal_Struct_Borsh              | 254.4     | 576   | 1         | (baseline for MarshalInto) |
| MarshalInto_Struct_Bin            | 123.7     | 0     | 0         | #2 EncodeInto |
| Marshal_Struct_Bin                | 178.1     | 576   | 1         | (baseline) |
| TxHeader_Cursor                   | 10.66     | 0     | 0         | #4 Cursor |
| TxHeader_Encoder                  | 65.21     | 112   | 1         | (baseline: Encoder-into) |
| TxHeader_Raw                      | 13.46     | 0     | 0         | (hand-rolled lower bound) |
| Cursor_8xU64LE                    | 4.14      | 0     | 0         | #4 Cursor |
| Encoder_8xU64LE                   | 48.69     | 112   | 1         | (baseline) |
| PatchBlockhash_ViewAs             | 0.23      | 0     | 0         | #7 ViewAs |
| PatchBlockhash_Copy               | 0.23      | 0     | 0         | (raw copy) |
| PatchBlockhash_DecodeEncode       | 180.7     | 128   | 2         | (no-ViewAs baseline) |

## Geomean

`geomean: 117.8 ns -> 80.3 ns` over all 36 shared benchmarks -- **-31.9% overall**.

## Techniques landed on this branch

| #  | Technique                                  | Headline delta |
| -- | ------------------------------------------ | -------------- |
| #2 | EncodeInto (pre-sized output buffer)       | 1 alloc -> 0 allocs; -16% to -28% ns/op |
| #8 | Bounded allocations (MaxSliceLen/MaxMapLen, element-size-aware checks) | Closes map DoS (2^32 -> error) and slice element-size amplification. Zero perf cost. |
| #4 | Cursor (zero-overhead write cursor)        | 6.8-11.7x faster than Encoder for hand-rolled encoders |
| #7 | ViewAs (in-place field mutation)           | 730x faster than decode-then-encode round-trip for patches |

## Reproducing

```sh
# On HEAD
go test -bench . -benchmem -benchtime=500ms -count=6 -run '^$' ./binary/ > /tmp/bench-head.out

# Checkout the initial branch commit in a worktree to capture the baseline
git worktree add --detach /tmp/solana-initial 56bb047
(cd /tmp/solana-initial && go test -bench . -benchmem -benchtime=500ms -count=6 -run '^$' ./binary/ > /tmp/bench-initial.out)
git worktree remove /tmp/solana-initial

# Compare
go install golang.org/x/perf/cmd/benchstat@latest
~/go/bin/benchstat /tmp/bench-initial.out /tmp/bench-head.out
```
