package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

var _ quarantine.MsgServer = Keeper{}

func (k Keeper) OptIn(goCtx context.Context, msg *quarantine.MsgOptIn) (*quarantine.MsgOptInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	k.SetOptIn(ctx, toAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgOptInResponse{}, nil
}

func (k Keeper) OptOut(goCtx context.Context, msg *quarantine.MsgOptOut) (*quarantine.MsgOptOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	k.SetOptOut(ctx, toAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgOptOutResponse{}, nil
}

func (k Keeper) Accept(goCtx context.Context, msg *quarantine.MsgAccept) (*quarantine.MsgAcceptResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}

	fromAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %v", err)
	}

	funds := k.GetQuarantineRecord(ctx, toAddr, fromAddr)
	if !funds.IsZero() {
		qHolderAddr := k.GetFundsHolder()
		if len(qHolderAddr) == 0 {
			return nil, sdkerrors.ErrUnknownAddress.Wrapf("no quarantine holder account defined")
		}

		if err = k.bankKeeper.SendCoinsBypassQuarantine(ctx, qHolderAddr, toAddr, funds.Coins); err != nil {
			return nil, err
		}
	}

	k.SetQuarantineRecordAccepted(ctx, toAddr, fromAddr)

	if msg.Permanent {
		k.SetAutoResponse(ctx, toAddr, fromAddr, quarantine.AUTO_RESPONSE_ACCEPT)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgAcceptResponse{}, nil
}

func (k Keeper) Decline(goCtx context.Context, msg *quarantine.MsgDecline) (*quarantine.MsgDeclineResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}

	fromAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %v", err)
	}

	k.SetQuarantineRecordDeclined(ctx, toAddr, fromAddr)

	if msg.Permanent {
		k.SetAutoResponse(ctx, toAddr, fromAddr, quarantine.AUTO_RESPONSE_DECLINE)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgDeclineResponse{}, nil
}

func (k Keeper) UpdateAutoResponses(goCtx context.Context, msg *quarantine.MsgUpdateAutoResponses) (*quarantine.MsgUpdateAutoResponsesResponse, error) {
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
		k.SetAutoResponse(ctx, toAddr, fromAddr, update.Response)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgUpdateAutoResponsesResponse{}, nil
}
