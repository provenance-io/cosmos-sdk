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

func (k Keeper) ImmediateSanction(ctx context.Context, req *sanction.MsgImmediateSanction) (*sanction.MsgImmediateSanctionResponse, error) {
	// TODO[1046]: Implement ImmediateSanction
	panic("not implemented")
}

func (k Keeper) ImmediateUnsanction(ctx context.Context, req *sanction.MsgImmediateUnsanction) (*sanction.MsgImmediateUnsanctionResponse, error) {
	// TODO[1046]: Implement ImmediateUnsanction
	panic("not implemented")
}
