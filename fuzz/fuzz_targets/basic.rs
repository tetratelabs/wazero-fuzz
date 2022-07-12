#![no_main]

use libfuzzer_sys::arbitrary::{Result, Unstructured};
use libfuzzer_sys::fuzz_target;
use wasm_smith::SwarmConfig;

fuzz_target!(|data: &[u8]| {
    // errors in `run` have to do with not enough input in `data`, which we
    // ignore here since it doesn't affect how we'd like to fuzz.
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
    config.max_memory_pages = 10;
    config.reference_types_enabled = false;

    let module = wasm_smith::Module::new(config.clone(), &mut u)?;
    let module_byte = module.to_bytes();

    unsafe {
        run_wazero(module_byte.as_ptr(), module_byte.len());
    }
    Ok(())
}

extern "C" {
    fn run_wazero(binary_ptr: *const u8, binary_size: usize);
}
