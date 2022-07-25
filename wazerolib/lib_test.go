package main

import (
	"context"
	"os"
	"testing"

	"github.com/tetratelabs/wazero"
)

// TestReRunFailedCase re-runs the failed case specified by WASM_BINARY_NAME in testdata directory.
func TestReRunFailedCase(t *testing.T) {
	binaryPath := os.Getenv("WASM_BINARY_PATH")

	wasmBin, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatal(err)
	}

	// Choose the context to use for function calls.
	ctx := context.Background()

	compiler := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigCompiler().WithWasmCore2())
	interpreter := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigInterpreter().WithWasmCore2())

	// Compile module.
	compiledCompiled, err := compiler.CompileModule(ctx, wasmBin, wazero.NewCompileConfig())
	if err != nil {
		t.Fatal(err)
	}

	interpreterCompiled, err := interpreter.CompileModule(ctx, wasmBin, wazero.NewCompileConfig())
	if err != nil {
		t.Fatal(err)
	}

	// Instantiate module.
	compilerMod, compilerInstErr := compiler.InstantiateModule(ctx, compiledCompiled, wazero.NewModuleConfig())
	interpreterMod, interpreterInstErr := interpreter.InstantiateModule(ctx, interpreterCompiled, wazero.NewModuleConfig())

	okToInvoke, err := ensureInstantiationError(compilerInstErr, interpreterInstErr)
	if err != nil {
		t.Fatal(err)
	}

	if okToInvoke {
		if err = ensureInvocationResultMatch(compilerMod, interpreterMod, interpreterCompiled.ExportedFunctions()); err != nil {
			t.Fatal(err)
		}
	}
}
