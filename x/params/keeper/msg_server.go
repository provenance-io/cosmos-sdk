package keeper

import (
	"context"
	sdkerrors "cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	types "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

type msgServer struct {
	Keeper
}

func (m msgServer) ParamChange(goCtx context.Context, msg *types.ParameterChangeProposal) (*types.ParamChangeResponse, error) {
	if m.Keeper.GetAuthority() != msg.FromAddress {
		return nil, errors.Wrapf(gov.ErrInvalidSigner, "expected %s got %s", m.Keeper.GetAuthority(), msg.FromAddress)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, c := range msg.Changes {
		ss, ok := m.Keeper.GetSubspace(c.Subspace)
		if !ok {
			//return sdkerrors.Wrap(proposal.ErrUnknownSubspace, c.Subspace)
		}

		m.Keeper.Logger(ctx).Info(
			fmt.Sprintf("attempt to set new parameter value; key: %s, value: %s", c.Key, c.Value),
		)

		if err := ss.Update(ctx, []byte(c.Key), []byte(c.Value)); err != nil {
			return sdkerrors.Wrapf(proposal.ErrSettingParameter, "key: %s, value: %s, err: %s", c.Key, c.Value, err.Error())
		}
	}

	return types.ParamChangeResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
