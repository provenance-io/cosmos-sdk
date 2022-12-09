package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

var _ sanction.MsgServer = Keeper{}

func (k Keeper) Sanction(ctx context.Context, req *sanction.MsgSanction) (*sanction.MsgSanctionResponse, error) {
	// TODO[1046]: Implement Sanction
	panic("not implemented")
}

func (k Keeper) Unsanction(ctx context.Context, req *sanction.MsgUnsanction) (*sanction.MsgUnsanctionResponse, error) {
	// TODO[1046]: Implement Unsanction
	panic("not implemented")
}

func (k Keeper) UpdateParams(ctx context.Context, req *sanction.MsgUpdateParams) (*sanction.MsgUpdateParamsResponse, error) {
	// TODO[1046]: Implement UpdateParams
	panic("not implemented")
}
