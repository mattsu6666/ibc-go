package types

import (
	"github.com/CosmWasm/wasmvm/api"
	"github.com/CosmWasm/wasmvm/types"
	ics23 "github.com/confio/ics23/go"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/ibc-go/modules/core/02-client/types"
	types3 "github.com/cosmos/ibc-go/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/modules/core/28-wasm/keeper"
	"github.com/cosmos/ibc-go/modules/core/exported"
)

const GasMultiplier uint64 = 100
const maxGasLimit = uint64(0x7FFFFFFFFFFFFFFF)

var _ exported.ClientState = (*ClientState)(nil)

type queryResponse struct {
	ProofSpecs      []*ics23.ProofSpec       `json:"proof_specs,omitempty"`
	Height          types2.Height            `json:"height,omitempty"`
	GenesisMetadata []types2.GenesisMetadata `json:"genesis_metadata,omitempty"`
	Result          contractResult           `json:"result,omitempty"`
	Root            types3.MerkleRoot        `json:"root,omitempty"`
	Timestamp       uint64                   `json:"timestamp,omitempty"`
	Status          exported.Status          `json:"status,omitempty"`
}

type contractResult struct {
	IsValid  bool   `json:"is_valid,omitempty"`
	ErrorMsg string `json:"err_msg,omitempty"`
}

type clientStateCallResponse struct {
	Me                *ClientState    `json:"me,omitempty"`
	NewConsensusState *ConsensusState `json:"new_consensus_state,omitempty"`
	NewClientState    *ClientState    `json:"new_client_state,omitempty"`
	Result            contractResult  `json:"result,omitempty"`
}

func (r *clientStateCallResponse) resetImmutables(c *ClientState) {
	if r.Me != nil {
		r.Me.CodeId = c.CodeId
	}
}

// Calls vm.Init with appropriate arguments
func initContract(codeID []byte, ctx sdk.Context, store sdk.KVStore, msg []byte) (*types.InitResponse, error) {
	gasMeter := ctx.GasMeter()
	chainID := ctx.BlockHeader().ChainID
	height := ctx.BlockHeader().Height
	// safety checks before casting below
	if height < 0 {
		panic("Block height must never be negative")
	}
	sec := ctx.BlockTime().Unix()
	if sec < 0 {
		panic("Block (unix) time must never be negative ")
	}
	nano := ctx.BlockTime().Nanosecond()
	env := types.Env{
		Block: types.BlockInfo{
			Height:    uint64(height),
			Time:      uint64(sec),
			TimeNanos: uint64(nano),
			ChainID:   chainID,
		},
		Contract: types.ContractInfo{
			Address: "",
		},
	}

	msgInfo := types.MessageInfo{
		Sender:    "",
		SentFunds: nil,
	}
	mockFailureAPI := *api.NewMockFailureAPI()
	mockQuerier := api.MockQuerier{}

	response, _, err := keeper.WasmVM.Instantiate(codeID, env, msgInfo, msg, store, mockFailureAPI, mockQuerier, gasMeter, gasMeter.Limit())
	return response, err
}

// Calls vm.Execute with internally constructed Gas meter and environment
func callContract(codeID []byte, ctx sdk.Context, store sdk.KVStore, msg []byte) (*types.HandleResponse, error) {
	gasMeter := ctx.GasMeter()
	chainID := ctx.BlockHeader().ChainID
	height := ctx.BlockHeader().Height
	// safety checks before casting below
	if height < 0 {
		panic("Block height must never be negative")
	}
	sec := ctx.BlockTime().Unix()
	if sec < 0 {
		panic("Block (unix) time must never be negative ")
	}
	nano := ctx.BlockTime().Nanosecond()
	env := types.Env{
		Block: types.BlockInfo{
			Height:    uint64(height),
			Time:      uint64(sec),
			TimeNanos: uint64(nano),
			ChainID:   chainID,
		},
		Contract: types.ContractInfo{
			Address: "",
		},
	}

	return callContractWithEnvAndMeter(codeID, &ctx, store, env, gasMeter, msg)
}

// Calls vm.Execute with supplied environment and gas meter
func callContractWithEnvAndMeter(codeID []byte, ctx *sdk.Context, store sdk.KVStore, env types.Env, gasMeter sdk.GasMeter, msg []byte) (*types.HandleResponse, error) {
	msgInfo := types.MessageInfo{
		Sender:    "",
		SentFunds: nil,
	}
	mockFailureAPI := *api.NewMockFailureAPI()
	mockQuerier := api.MockQuerier{}

	resp, gasUsed, err := keeper.WasmVM.Execute(codeID, env, msgInfo, msg, store, mockFailureAPI, mockQuerier, gasMeter, gasMeter.Limit())
	if ctx != nil {
		consumeGas(*ctx, gasUsed)
	}
	return resp, err
}

func queryContractWithStore(codeID []byte, store sdk.KVStore, msg []byte) ([]byte, error) {
	mockEnv := api.MockEnv()
	mockGasMeter := api.NewMockGasMeter(1)
	mockFailureAPI := *api.NewMockFailureAPI()
	mockQuerier := api.MockQuerier{}

	resp, _, err := keeper.WasmVM.Query(codeID, mockEnv, msg, store, mockFailureAPI, mockQuerier, mockGasMeter, maxGasLimit)
	return resp, err
}

func consumeGas(ctx sdk.Context, gas uint64) {
	consumed := gas / GasMultiplier
	ctx.GasMeter().ConsumeGas(consumed, "wasm contract")
	// throw OutOfGas error if we ran out (got exactly to zero due to better limit enforcing)
	if ctx.GasMeter().IsOutOfGas() {
		panic(sdk.ErrorOutOfGas{Descriptor: "Wasmer function execution"})
	}
}
