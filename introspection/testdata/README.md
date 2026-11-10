# Fixtures

`getTransaction` results captured from mainnet on 2026-11-09. Each
`<name>.<encoding>.json` file holds the `result` of the JSON-RPC response.

| name | signature | slot | version | covers |
|---|---|---|---|---|
| `sync_native` | `3wMPKe86EK4zZEuD6GycPoi3CvWrPBvvwzdLcrLv6JW2FNSq89Re1jhPbSvnw9H54aAMkkeU5F72k21mgAjYwqgB` | 175687414 | legacy | Wrap SOL: a System transfer, then a Token SyncNative (upstream issue #153). |
| `swap_v0` | `2gyoib4jMXPiDWEjdXF7aGyZoFpAVuYtqi8vzowyMuAcxvJS9979Lm1WxDzhA4NMCrdo2i1gVMH6tHnWgtRRnjFg` | 467421519 | 0 | Jupiter swap: 2 lookup tables (5 writable and 14 readonly loaded accounts), CPI depth 3, top-level instructions using static, loaded writable and loaded readonly accounts. |
| `failed_v0` | `3o1pyAxPBCwYvUriuVJzYmbpSTKZijBxWYMixRgLrwdgHz7tMKgMfqN8MzHR3TUkFeioUY4yr5bhMg2sBv1SVtp9` | 467421524 | 0 | Failed Jupiter swap (custom error 6001 in instruction 3), with the inner instructions recorded up to the failure, and 3 lookup tables. |
| `swap_v1` | `Vy2DvACx87dQHMoFBeDyfpANWtcxSG1H8b6PKHKfuBqbN3TX6ktmoGtLCYRAjgzsMjTduFsbnESgfC4TPhCeE6N` | 467421524 | 1 | v1 (SIMD-0385) swap with a transaction config and CPI depth 3. |

The RPC serves base58 only with `"maxSupportedTransactionVersion": 0`, so
the v1 transaction has no base58 file. The jsonParsed files are the oracle
of `TestWalk_MatchesAgaveJSONParsed`: they hold the account roles and raw
instructions as resolved by agave itself.

To capture a file again (use `"maxSupportedTransactionVersion": 0` for
base58):

```sh
curl -s https://api.mainnet-beta.solana.com -H 'Content-Type: application/json' -d '{
  "jsonrpc": "2.0", "id": 1, "method": "getTransaction",
  "params": ["<signature>", {"encoding": "<encoding>", "maxSupportedTransactionVersion": 1, "commitment": "confirmed"}]
}' | jq -c .result > <name>.<encoding>.json
```
