package keeper

import (
	"context"
	"encoding/hex"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	wasm "github.com/CosmWasm/wasmvm"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ibc-go/v3/modules/core/28-wasm/types"
)

var _ types.QueryServer = (*Keeper)(nil)

func (q Keeper) WasmCode(c context.Context, query *types.WasmCodeQuery) (*types.WasmCodeResponse, error) {
	if query == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	rawID, err := hex.DecodeString(query.CodeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid code id")
	}
	codeID := wasm.Checksum(rawID)
	if WasmVM == nil {
		return nil, status.Error(codes.Internal, "WasmVM resource not initialized")
	}
	code, err := WasmVM.GetCode(codeID)
	if err != nil {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrap(types.ErrWasmCodeIDNotFound, query.CodeId).Error(),
		)
	}
	return &types.WasmCodeResponse{
		Code: code,
	}, nil
}
