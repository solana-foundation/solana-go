# Changelog

## [1.19.1](https://github.com/solana-foundation/solana-go/compare/v1.19.0...v1.19.1) (2026-04-26)


### Performance Improvements

* migrate to curve25519-voi for ed25519 operations ([20713fb](https://github.com/solana-foundation/solana-go/commit/20713fbbe52d4d20cab792a702771790346f19c7))

## [1.19.0](https://github.com/solana-foundation/solana-go/compare/v1.18.0...v1.19.0) (2026-04-23)


### Features

* is token mint classifier ([4f72982](https://github.com/solana-foundation/solana-go/commit/4f72982442c9b3c166b72dbb2de730f58b575539))


### Bug Fixes

* enhance getUint64 function to handle string inputs ([5309095](https://github.com/solana-foundation/solana-go/commit/53090952ffc598c1870617b1727179135994ec65))
* keep websocket request IDs within JSON-safe range ([8ed3105](https://github.com/solana-foundation/solana-go/commit/8ed31050f7af62f26b5615f40546bb498cab9219))
* **message:** json version detection ([1fd2201](https://github.com/solana-foundation/solana-go/commit/1fd2201431de71d9164d281eef2c62f858fb5016))
* **message:** use gojson ([8d211d5](https://github.com/solana-foundation/solana-go/commit/8d211d5dc9e610b54fb84f662d83e2f55668e9d4))
* reject malformed ed25519 private keys in PrivateKeyFromBase58 ([edcedcc](https://github.com/solana-foundation/solana-go/commit/edcedcc2ba5ebd01c65baf64b8a22bf879cb0d55))
* **rpc:** match ParsedTransactionMeta to TransactionMeta ([a0f95c2](https://github.com/solana-foundation/solana-go/commit/a0f95c23eac6031c0f44e3095b763da531b8b2b7))

### Performance Improvements

* **json:** swap encoding/json and jsoniter for goccy/go-json ([c445f76](https://github.com/solana-foundation/solana-go/commit/c445f76c249d944731983fd720c2a9e6a874dc62))
* **transaction:** add cap hints and use pk instead of str ([91e8cec](https://github.com/solana-foundation/solana-go/commit/91e8cec9785fccd2663f28e61c8cc5353f38c419))


## [1.18.0](https://github.com/solana-foundation/solana-go/compare/v1.17.0...v1.18.0) (2026-04-16)


### Features

* add getters to txn with meta
* add token-2022 extensions 
* stake state types & ext tests 
* vote program complete 

### Bug Fixes

* allign rpc client with agave 
* memo program parity 

### Performance Improvements

* **message:** eliminate complex scans, struct copies, and redundant allocs
