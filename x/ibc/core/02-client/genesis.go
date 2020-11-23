package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/09-localhost/types"
)

// InitGenesis initializes the ibc client submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	for _, client := range gs.Clients {
		cs, ok := client.ClientState.GetCachedValue().(exported.ClientState)
		if !ok {
			panic("invalid client state")
		}

		k.SetClientState(ctx, client.ClientId, cs)
	}

	// Set all client metadata
	if len(gs.ClientsMetadata) != 0 {
		k.SetAllClientMetadata(ctx, gs.ClientsMetadata)
	}

	for _, cs := range gs.ClientsConsensus {
		for _, consState := range cs.ConsensusStates {
			consensusState, ok := consState.ConsensusState.GetCachedValue().(exported.ConsensusState)
			if !ok {
				panic(fmt.Sprintf("invalid consensus state with client ID %s at height %s", cs.ClientId, consState.Height))
			}

			k.SetClientConsensusState(ctx, cs.ClientId, consState.Height, consensusState)
		}
	}

	if !gs.CreateLocalhost {
		return
	}

	// NOTE: return if the localhost client was already imported. The chain-id and
	// block height will be overwriten to the correct values during BeginBlock.
	if _, found := k.GetClientState(ctx, exported.Localhost); found {
		return
	}

	// client id is always "localhost"
	version := types.ParseChainID(ctx.ChainID())
	clientState := localhosttypes.NewClientState(
		ctx.ChainID(), types.NewHeight(version, uint64(ctx.BlockHeight())),
	)

	if err := k.CreateClient(ctx, exported.Localhost, clientState, nil); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the ibc client submodule's exported genesis.
// NOTE: CreateLocalhost should always be false on export since a
// created localhost will be included in the exported clients.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	genClients := k.GetAllGenesisClients(ctx)
	clientsMetadata, err := k.GetAllClientMetadata(ctx, genClients)
	if err != nil {
		panic(err)
	}
	return types.GenesisState{
		Clients:          genClients,
		ClientsMetadata:  clientsMetadata,
		ClientsConsensus: k.GetAllConsensusStates(ctx),
		CreateLocalhost:  false,
	}
}