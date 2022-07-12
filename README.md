## wazero-fuzz

Fuzzing infrastructure for [wazero](https://github.com/tetratelabs/wazero) via [wasm-tools](https://github.com/bytecodealliance/wasm-tools)
and [libFuzzer](https://llvm.org/docs/LibFuzzer.html).

### Dependency

- [cargo](https://doc.rust-lang.org/cargo/getting-started/installation.html)
    Needs to enable nightly (for libFuzzer).
- [cargo-fuzz](https://github.com/rust-fuzz/cargo-fuzz)
  - `cargo install cargo-fuzz`

### Run

```
cargo fuzz run basic
```
