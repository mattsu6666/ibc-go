package keeper

import (
	"strings"

	wasm "github.com/CosmWasm/wasmvm"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/cosmos/ibc-go/v3/modules/core/28-wasm/types"
	"github.com/tendermint/tendermint/libs/log"
)

// WasmVM initialized by wasm keeper
var WasmVM *wasm.VM

// VMConfig represents Wasm virtual machine settings
type VMConfig struct {
	DataDir           string
	SupportedFeatures []string
	MemoryLimitMb     uint32
	PrintDebug        bool
	CacheSizeMb       uint32
}

// Keeper will have a reference to Wasmer with it's own data directory.
type Keeper struct {
    storeKey      sdk.StoreKey
	cdc           codec.BinaryCodec
	wasmValidator *WasmValidator
}

func NewKeeper(cdc codec.BinaryCodec, key sdk.StoreKey, vmConfig *VMConfig, validationConfig *ValidationConfig) Keeper {
	supportedFeatures := strings.Join(vmConfig.SupportedFeatures, ",")
	vm, err := wasm.NewVM(vmConfig.DataDir, supportedFeatures, vmConfig.MemoryLimitMb, vmConfig.PrintDebug, vmConfig.CacheSizeMb)
	if err != nil {
		panic(err)
	}
	wasmValidator, err := NewWasmValidator(validationConfig, func() (*wasm.VM, error) { return wasm.NewVM(vmConfig.DataDir, supportedFeatures, vmConfig.MemoryLimitMb, vmConfig.PrintDebug, vmConfig.CacheSizeMb)
	})
	if err != nil {
		panic(err)
	}
	WasmVM = vm
	return Keeper{
		cdc:           cdc,
        storeKey:      key,
		wasmValidator: wasmValidator,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+host.ModuleName+"/"+types.SubModuleName)
}

func (k Keeper) PushNewWasmCode(ctx sdk.Context, code []byte) (wasm.Checksum, error) {
	codeID, err := WasmVM.Create(code)
	if err != nil {
		k.Logger(ctx).Error("in creating new wasmcode", err)
		return nil, types.ErrWasmInvalidCode
	}
	if isValidWasmCode, err := k.wasmValidator.validateWasmCode(code); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrWasmCodeValidation, "unable to validate wasm code: %s", err)
	} else if !isValidWasmCode {
		return nil, types.ErrWasmInvalidCode
	}
	return codeID, nil
}
