#![no_main]

use libfuzzer_sys::arbitrary::{Result, Unstructured};
use libfuzzer_sys::fuzz_target;
use wasm_smith::SwarmConfig;

fuzz_target!(|data: &[u8]| {
    drop(run(data));
});

fn run(data: &[u8]) -> Result<()> {
    // Create the random source.
    let mut u = Unstructured::new(data);

    // Generate the configuration.
    let mut config: SwarmConfig = u.arbitrary()?;

    // 64-bit memory won't be supported by wazero.
    config.memory64_enabled = false;
    // Also, multiple memories are not supported.
    config.max_memories = 1;
    // TODO: after having CompiledModule.Imports(), enable this with dummy imports.
    // See https://github.com/bytecodealliance/wasmtime/blob/f242975c49385edafe4f72dfa5f0ff6aae23eda3/crates/fuzzing/src/oracles/dummy.rs#L6-L20
    config.max_imports = 0;
    // If we don't set the limit, we will soon reach the OOM and the fuzzing will be killed by OS.
    config.max_memory_pages = 10;

    // Generate the random module via wasm-smith.
    let module_byte = wasm_smith::Module::new(config.clone(), &mut u)?.to_bytes();

    // Pass the randomly generated module to the wazero library.
    unsafe {
        run_wazero(module_byte.as_ptr(), module_byte.len());
    }

    // We always return Ok as inside of run_wazero, we cause panic if the binary is interesting.
    Ok(())
}

extern "C" {
    // run_wazero is implemented in Go, and accepts the pointer to the binary and its size.
    fn run_wazero(binary_ptr: *const u8, binary_size: usize);
}
