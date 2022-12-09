package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

var _ sanction.QueryServer = Keeper{}

func (k Keeper) IsSanctioned(goCtx context.Context, req *sanction.QueryIsSanctionedRequest) (*sanction.QueryIsSanctionedResponse, error) {
	// TODO[1046]: Implement IsSanctioned
	panic("not implemented")
}

func (k Keeper) SanctionedAddresses(goCtx context.Context, req *sanction.QuerySanctionedAddressesRequest) (*sanction.QuerySanctionedAddressesResponse, error) {
	// TODO[1046]: Implement SanctionedAddresses
	panic("not implemented")
}

func (k Keeper) Params(goCtx context.Context, _ *sanction.QueryParamsRequest) (*sanction.QueryParamsResponse, error) {
	// TODO[1046]: Implement Params
	panic("not implemented")
}
