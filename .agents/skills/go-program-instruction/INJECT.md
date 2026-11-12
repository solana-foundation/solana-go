---
heading: Go Program Instructions
---

- Read the Rust definition of the instruction first: its tag, its data layout and its account list. Then copy the shape of a neighboring instruction in the same package.
- Register a new instruction everywhere in `instructions.go`: the id constant, `InstructionIDToName` and the `InstructionImplDef` variant list. Never insert an id in the middle of an `iota` block.
- The account order and the signer and writable flags must match the Rust docs exactly.
- Test the exact bytes of `Data()` and decode those bytes back with `DecodeInstruction`. The round trip test is not enough.
