package types

import (
	"encoding/hex"
	"fmt"
	wasmvm "github.com/CosmWasm/wasmvm"
)

const (
	SubModuleName = "wasm-manager"
)

func CodeID(codeID wasmvm.Checksum) []byte {
	return []byte(fmt.Sprintf("code_id/%s", hex.EncodeToString(codeID)))
}
