package keeper

import (
	"context"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) Send(goCtx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.IsSendEnabledCoins(ctx, msg.Amount...); err != nil {
		return nil, err
	}

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if k.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	err = k.SendCoins(ctx, from, to, msg.Amount)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, a := range msg.Amount {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "send"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgSendResponse{}, nil
}

func (k msgServer) MultiSend(goCtx context.Context, msg *types.MsgMultiSend) (*types.MsgMultiSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// NOTE: totalIn == totalOut should already have been checked
	for _, in := range msg.Inputs {
		if err := k.IsSendEnabledCoins(ctx, in.Coins...); err != nil {
			return nil, err
		}
	}

	for _, out := range msg.Outputs {
		accAddr := sdk.MustAccAddressFromBech32(out.Address)

		if k.BlockedAddr(accAddr) {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive transactions", out.Address)
		}
	}

	err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgMultiSendResponse{}, nil
}

func (k msgServer) QuarantineOptIn(goCtx context.Context, msg *types.MsgQuarantineOptIn) (*types.MsgQuarantineOptInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	k.SetQuarantineOptIn(ctx, toAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgQuarantineOptInResponse{}, nil
}

func (k msgServer) QuarantineOptOut(goCtx context.Context, msg *types.MsgQuarantineOptOut) (*types.MsgQuarantineOptOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	k.SetQuarantineOptOut(ctx, toAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgQuarantineOptOutResponse{}, nil
}

func (k msgServer) QuarantineAccept(goCtx context.Context, msg *types.MsgQuarantineAccept) (*types.MsgQuarantineAcceptResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}

	fromAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %v", err)
	}

	funds := k.GetQuarantinedFunds(ctx, toAddr, fromAddr)
	if !funds.IsZero() {
		if err = k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, funds.Coins); err != nil {
			return nil, err
		}
	}

	k.SetQuarantinedFundsAccepted(ctx, toAddr, fromAddr)

	if msg.Permanent {
		k.SetQuarantineAutoResponse(ctx, toAddr, fromAddr, types.QUARANTINE_AUTO_RESPONSE_ACCEPT)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgQuarantineAcceptResponse{}, nil
}

func (k msgServer) QuarantineDecline(goCtx context.Context, msg *types.MsgQuarantineDecline) (*types.MsgQuarantineDeclineResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}

	fromAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %v", err)
	}

	k.SetQuarantinedFundsDeclined(ctx, toAddr, fromAddr)

	if msg.Permanent {
		k.SetQuarantineAutoResponse(ctx, toAddr, fromAddr, types.QUARANTINE_AUTO_RESPONSE_DECLINE)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgQuarantineDeclineResponse{}, nil
}

func (k msgServer) UpdateQuarantineAutoResponses(goCtx context.Context, msg *types.MsgUpdateQuarantineAutoResponses) (*types.MsgUpdateQuarantineAutoResponsesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}

	for _, update := range msg.Updates {
		fromAddr, err := sdk.AccAddressFromBech32(update.FromAddress)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %v", err)
		}
		k.SetQuarantineAutoResponse(ctx, toAddr, fromAddr, update.Response)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgUpdateQuarantineAutoResponsesResponse{}, nil
}
