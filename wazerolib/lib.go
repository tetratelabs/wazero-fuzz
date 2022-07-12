package main

import "C"
import (
	"context"
	"log"
	"reflect"
	"unsafe"

	"github.com/tetratelabs/wazero"
)

func main() {}

//export run_wazero
func run_wazero(binaryPtr uintptr, binarySize int) {
	wasmBin := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: binaryPtr,
		Len:  binarySize,
		Cap:  binarySize,
	}))

	// Choose the context to use for function calls.
	ctx := context.Background()

	runtimes := []wazero.Runtime{
		wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigCompiler().WithWasmCore2()),
		wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigInterpreter().WithWasmCore2()),
	}

	defer runtimes[0].Close(ctx)
	defer runtimes[1].Close(ctx)

	for _, r := range runtimes {
		_, err := r.InstantiateModuleFromBinary(ctx, wasmBin)
		if err != nil {
			log.Panicln(err)
		}

		// Invokes all the functions.
		//fmt.Println(mod)
	}
	return
}
