# Changelog

## [1.22.0](https://github.com/solana-foundation/solana-go/compare/v1.21.0...v1.22.0) (2026-06-30)


### Features

* add nonce account support ([#456](https://github.com/solana-foundation/solana-go/issues/456)) ([336881c](https://github.com/solana-foundation/solana-go/commit/336881cfbe66902ee0ec07b25dc1912186508a72))
* add sysvars ([#457](https://github.com/solana-foundation/solana-go/issues/457)) ([f777896](https://github.com/solana-foundation/solana-go/commit/f7778964b61f6ac2622a194fb034dd92da59c67f))
* **address-lookup-table:** one-call resolve for message lookups (closes [#262](https://github.com/solana-foundation/solana-go/issues/262)) ([#445](https://github.com/solana-foundation/solana-go/issues/445)) ([34beab1](https://github.com/solana-foundation/solana-go/commit/34beab1e231b6e85b7232361997d70bdd79825ef))
* **rpc:** add getTransactionsForAddress client method (closes [#343](https://github.com/solana-foundation/solana-go/issues/343)) ([#450](https://github.com/solana-foundation/solana-go/issues/450)) ([9e538c8](https://github.com/solana-foundation/solana-go/commit/9e538c84246ef1a3abfbcc96cac7a3196889b479))
* **rpc:** forward minContextSlot in 4 remaining JSON-RPC endpoints ([#448](https://github.com/solana-foundation/solana-go/issues/448)) ([c176402](https://github.com/solana-foundation/solana-go/commit/c176402c339e25c704f4eba9be540d56f770feb0))
* **rpc:** forward MinContextSlot in getBalance/getLatestBlockhash/getSlot/getTokenAccountBalance ([#442](https://github.com/solana-foundation/solana-go/issues/442)) ([b8e70e8](https://github.com/solana-foundation/solana-go/commit/b8e70e8c5cdd1229c6126f00090527664aa8496c))
* **zk:** add ElGamal & AES key derivation ([#413](https://github.com/solana-foundation/solana-go/issues/413)) ([9fcbf0c](https://github.com/solana-foundation/solana-go/commit/9fcbf0ced8af5b5ad975e45999478b5c36c3e65a))


### Bug Fixes

* **message:** surface typed `ErrAddressTablesNotSet` from AccountMetaList (closes [#280](https://github.com/solana-foundation/solana-go/issues/280)) ([#441](https://github.com/solana-foundation/solana-go/issues/441)) ([a87922d](https://github.com/solana-foundation/solana-go/commit/a87922db64914ec510f0bd3994fa97f9e8cb41cc))
* **rpc:** default simulateTransaction Accounts.Encoding to base64 (closes [#446](https://github.com/solana-foundation/solana-go/issues/446)) ([#447](https://github.com/solana-foundation/solana-go/issues/447)) ([2697614](https://github.com/solana-foundation/solana-go/commit/26976146120a5030f35cd108d6a1257554c51cfa))
* **token,token-2022:** Build() sets Impl to *T, matching DecodeInstruction (closes [#222](https://github.com/solana-foundation/solana-go/issues/222)) ([#440](https://github.com/solana-foundation/solana-go/issues/440)) ([38c57db](https://github.com/solana-foundation/solana-go/commit/38c57dbfbc5f1b18fffa9bbc029f867f8a9116c5))
* **token,token-2022:** per-instruction ProgramID override (closes [#254](https://github.com/solana-foundation/solana-go/issues/254)) ([#439](https://github.com/solana-foundation/solana-go/issues/439)) ([f489aac](https://github.com/solana-foundation/solana-go/commit/f489aaced7390fc7bd0fb2ceed2773f2414992d7))
* **ws:** surface subscription request errors to Recv (closes [#175](https://github.com/solana-foundation/solana-go/issues/175)) ([#449](https://github.com/solana-foundation/solana-go/issues/449)) ([725147a](https://github.com/solana-foundation/solana-go/commit/725147a6691fa580d392ec7ccf844da41d4ddafe))

## [1.21.0](https://github.com/solana-foundation/solana-go/compare/v1.20.0...v1.21.0) (2026-05-25)


### Features

* **rpc:** add NewWithCommitment / NewWithTimeout / NewWithTimeoutAndCommitment ([#436](https://github.com/solana-foundation/solana-go/issues/436)) ([e93ff5e](https://github.com/solana-foundation/solana-go/commit/e93ff5e937733daca5fed01c362961c4c8aead25)), closes [#414](https://github.com/solana-foundation/solana-go/issues/414)
* **rpc:** forward MinContextSlot in getProgramAccounts and getTokenAccounts ([#431](https://github.com/solana-foundation/solana-go/issues/431)) ([17984a5](https://github.com/solana-foundation/solana-go/commit/17984a55c17ab0fc9f308872a43b737601d6a8da))
* **rpc:** support EncodingJSON in GetTransaction ([#420](https://github.com/solana-foundation/solana-go/issues/420)) ([b906b70](https://github.com/solana-foundation/solana-go/commit/b906b70527a5dfed358090e27dd7f4a7f12749c3))
* **wallet:** derive PrivateKey/Wallet from BIP-39 mnemonic ([#429](https://github.com/solana-foundation/solana-go/issues/429)) ([89ef706](https://github.com/solana-foundation/solana-go/commit/89ef706472ad49a9622a058497852711f7bd3771))
* **ws:** support dataSlice in AccountSubscribe ([#433](https://github.com/solana-foundation/solana-go/issues/433)) ([fb31fb1](https://github.com/solana-foundation/solana-go/commit/fb31fb13b42141bb6067c7447b8618e7e848b97b))
* **ws:** support dataSlice in ProgramSubscribe ([#434](https://github.com/solana-foundation/solana-go/issues/434)) ([950b110](https://github.com/solana-foundation/solana-go/commit/950b110b8f369de33143705cfba0b8da7d240d6f))
* **ws:** support enableReceivedNotification in SignatureSubscribe ([#432](https://github.com/solana-foundation/solana-go/issues/432)) ([810f171](https://github.com/solana-foundation/solana-go/commit/810f171ff933c1508e9526a2a536a287cac7c386))


### Bug Fixes

* **rpc:** support EncodingJSON in GetBlockWithOpts ([#419](https://github.com/solana-foundation/solana-go/issues/419)) ([eee363a](https://github.com/solana-foundation/solana-go/commit/eee363a738642efc6006cdce863689d49afc712c))
* **ws:** reject EncodingJSONParsed in BlockSubscribe ([#426](https://github.com/solana-foundation/solana-go/issues/426)) ([bf130a2](https://github.com/solana-foundation/solana-go/commit/bf130a2a69b0a3f0462f8119c6b03dd1e9282cf8))
* **ws:** use spec "showRewards" key in blockSubscribe params ([#430](https://github.com/solana-foundation/solana-go/issues/430)) ([6969f12](https://github.com/solana-foundation/solana-go/commit/6969f121e5700803befeb089e9dc4bbecfdb5f89))
* **ws:** use uint64 for params.Subscription in incoming notifications ([#427](https://github.com/solana-foundation/solana-go/issues/427)) ([427de1a](https://github.com/solana-foundation/solana-go/commit/427de1a9f438b658dd649ba6f13ba81558192ee1))

## [1.20.0](https://github.com/solana-foundation/solana-go/compare/v1.19.0...v1.20.0) (2026-05-08)


### Features

* **jsonrpc:** add CustomHeader http.Header for multi-value headers ([20b37ba](https://github.com/solana-foundation/solana-go/commit/20b37ba403c438ebe914b43ff7081f9598832d0c))


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
