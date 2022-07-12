package main

import "C"
import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/tetratelabs/wazero"
)

func main() {}

//export run_wazero
//
// run_wazero ensures that the behavior is the same between the compiler and the interpreter for any given
// binary.
func run_wazero(binaryPtr uintptr, binarySize int) {
	wasmBin := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: binaryPtr,
		Len:  binarySize,
		Cap:  binarySize,
	}))

	// Choose the context to use for function calls.
	ctx := context.Background()

	compiler := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigCompiler().WithWasmCore2())
	interpreter := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigInterpreter().WithWasmCore2())

	var failed = true
	defer func() {
		if failed {
			saveFailedBinary(wasmBin)
		}
	}()

	compilerRes, compilerErr := run(ctx, compiler, wasmBin)
	interpreterRes, interpreterErr := run(ctx, interpreter, wasmBin)

	if compilerErr != interpreterErr {
		panic(fmt.Sprintf("error mismatch: compiler got: '%v', but interpreter got '%v'\n", compilerErr, interpreterErr))
	}

	if len(compilerRes) != len(interpreterRes) {
		panic(fmt.Sprintf("result length mismatch: compiler got %d results, but interpreter %d results\n", len(compilerRes), len(interpreterRes)))
	}

	for i, cr := range compilerRes {
		ir := interpreterRes[i]
		if cr != ir {
			panic(fmt.Sprintf("result mismatch: compiler got %v, but interpreter got %v\n", compilerRes, interpreterRes))
		}
	}

	failed = false
	return
}

func run(ctx context.Context, r wazero.Runtime, bin []byte) (result []uint64, err error) {
	defer func() {
		err = r.Close(ctx)
	}()

	_, err = r.InstantiateModuleFromBinary(ctx, bin)
	if err != nil {
		return
	}

	// TODO: Invokes all the functions.

	return
}

const failedCasesDir = "wazerolib/testdata"

func saveFailedBinary(bin []byte) {
	checksum := sha256.Sum256(bin)
	checkSumStr := hex.EncodeToString(checksum[:])

	path := fmt.Sprintf("%s/%s.wasm", failedCasesDir, checkSumStr)
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	_, err = f.Write(bin)
	if err != nil {
		panic(err)
	}

	fmt.Printf(`

Failed Wasm binary has been written as %[1]s.wasm
To reproduce the failure, execute: WASM_BINARY_NAME=%[1]s.wasm go test wazerolib/...

`, checkSumStr)
}
