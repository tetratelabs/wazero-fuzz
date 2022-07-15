package main

import "C"
import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"unsafe"

	"github.com/tetratelabs/wazero"
)

func main() {}

func allowedErrorDuringInstantiation(errMsg string) bool {
	if strings.HasPrefix(errMsg, "data[") && strings.HasSuffix(errMsg, "]: out of bounds memory access") {
		return true
	}

	if strings.HasPrefix(errMsg, "start function[") && strings.Contains(errMsg, "failed: wasm error:") {
		return true
	}
	return false
}

//export run_wazero
//
// run_wazero ensures that the behavior is the same between the compiler and the interpreter for any given
// binary.
func run_wazero(binaryPtr uintptr, binarySize int, watPtr uintptr, watSize int) {
	wasmBin := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: binaryPtr,
		Len:  binarySize,
		Cap:  binarySize,
	}))

	wat := *(*string)(unsafe.Pointer(&reflect.SliceHeader{
		Data: watPtr,
		Len:  watSize,
		Cap:  watSize,
	}))

	// Choose the context to use for function calls.
	ctx := context.Background()

	// Create two runtimes.
	interpreter := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigInterpreter().WithWasmCore2())
	compiler := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigCompiler().WithWasmCore2())

	defer compiler.Close(ctx)
	defer interpreter.Close(ctx)

	var failed = true
	defer func() {
		if failed {
			saveFailedBinary(wasmBin, wat)
		}
	}()

	// Instantiate module.
	_, compilerInstErr := compiler.InstantiateModuleFromBinary(ctx, wasmBin)
	_, interpreterInstErr := interpreter.InstantiateModuleFromBinary(ctx, wasmBin)

	err := ensureInstantiationError(compilerInstErr, interpreterInstErr)
	if err != nil {
		panic(err)
	}

	failed = false
	return
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

const failedCasesDir = "wazerolib/testdata"

func saveFailedBinary(bin []byte, wat string) {
	checksum := sha256.Sum256(bin)
	checkSumStr := hex.EncodeToString(checksum[:])

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	testDataDir := path.Join(dir, failedCasesDir)
	binaryPath := path.Join(testDataDir, fmt.Sprintf("%s.wasm", checkSumStr))
	f, err := os.Create(binaryPath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	_, err = f.Write(bin)
	if err != nil {
		panic(err)
	}

	watPath := path.Join(testDataDir, fmt.Sprintf("%s.wat", checkSumStr))
	watF, err := os.Create(watPath)
	if err != nil {
		panic(err)
	}

	defer watF.Close()

	_, err = watF.Write([]byte(wat))
	if err != nil {
		panic(err)
	}

	fmt.Printf(`

Failed Wasm binary has been written to %s 
Failed Wasm Text has been written to %s
To reproduce the failure, execute: WASM_BINARY_PATH=%s go test wazerolib/...

Failed WebAssembly Text:
%s

`, binaryPath, watPath, binaryPath, wat)
}
