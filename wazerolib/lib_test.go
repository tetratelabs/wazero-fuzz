package main_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
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

	// Instantiate module.
	_, compilerInstErr := compiler.InstantiateModuleFromBinary(ctx, wasmBin)
	_, interpreterInstErr := interpreter.InstantiateModuleFromBinary(ctx, wasmBin)

	err = ensureInstantiationError(compilerInstErr, interpreterInstErr)
	if err != nil {
		t.Fatal(err)
	}
}

func ensureInstantiationError(compilerErr, interpErr error) error {
	if compilerErr == nil && interpErr == nil {
		return nil
	} else if compilerErr == nil && interpErr != nil {
		return fmt.Errorf("compiler returned no error, but interpreter got: %w", interpErr)
	} else if compilerErr != nil && interpErr == nil {
		return fmt.Errorf("interpreter returned no error, but compiler got: %w", compilerErr)
	}

	compilerErrMsg, interpErrMsg := compilerErr.Error(), interpErr.Error()
	if idx := strings.Index(compilerErrMsg, "\n"); idx >= 0 {
		compilerErrMsg = compilerErrMsg[:strings.Index(compilerErrMsg, "\n")]
	}
	if idx := strings.Index(interpErrMsg, "\n"); idx >= 0 {
		interpErrMsg = interpErrMsg[:strings.Index(interpErrMsg, "\n")]
	}

	if !allowedErrorDuringInstantiation(compilerErrMsg) {
		return fmt.Errorf("invalid erro occur with compiler: %v", compilerErr)
	} else if !allowedErrorDuringInstantiation(interpErrMsg) {
		return fmt.Errorf("invalid erro occur with interpreter: %v", interpErrMsg)
	}

	if compilerErrMsg != interpErrMsg {
		return fmt.Errorf("error mismatch:\n\tinterpreter: %v\n\tcompiler: %v", interpErr, compilerErr)
	}
	return nil
}

func allowedErrorDuringInstantiation(errMsg string) bool {
	if strings.HasPrefix(errMsg, "data[") && strings.HasSuffix(errMsg, "]: out of bounds memory access") {
		return true
	}

	if strings.HasPrefix(errMsg, "start function[") && strings.Contains(errMsg, "failed: wasm error:") {
		return true
	}
	return false
}
