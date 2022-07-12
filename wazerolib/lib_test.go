package main

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/tetratelabs/wazero"
)

// TestReRunFailedCase re-runs the failed case specified by WASM_BINARY_NAME in testdata directory.
func TestReRunFailedCase(t *testing.T) {
	binaryHash := os.Getenv("WASM_BINARY_NAME")

	wasmBin, err := os.ReadFile(path.Join("testdata", binaryHash))
	if err != nil {
		t.Fatal(err)
	}

	// Choose the context to use for function calls.
	ctx := context.Background()

	compiler := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigCompiler().WithWasmCore2())
	interpreter := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigInterpreter().WithWasmCore2())

	compilerRes, compilerErr := run(ctx, compiler, wasmBin)
	interpreterRes, interpreterErr := run(ctx, interpreter, wasmBin)

	if compilerErr != interpreterErr {
		t.Fatalf("error mismatch: compiler got: '%v', but interpreter got '%v'\n", compilerErr, interpreterErr)
	}

	if len(compilerRes) != len(interpreterRes) {
		t.Fatalf("result length mismatch: compiler got %d results, but interpreter %d results\n", len(compilerRes), len(interpreterRes))
	}

	for i, cr := range compilerRes {
		ir := interpreterRes[i]
		if cr != ir {
			t.Fatalf("result mismatch: compiler got %v, but interpreter got %v\n", compilerRes, interpreterRes)
		}
	}
}
