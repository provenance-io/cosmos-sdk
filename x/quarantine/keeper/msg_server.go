package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

var _ quarantine.MsgServer = Keeper{}

func (k Keeper) QuarantineOptIn(goCtx context.Context, msg *quarantine.MsgQuarantineOptIn) (*quarantine.MsgQuarantineOptInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	k.SetQuarantineOptIn(ctx, toAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgQuarantineOptInResponse{}, nil
}

func (k Keeper) QuarantineOptOut(goCtx context.Context, msg *quarantine.MsgQuarantineOptOut) (*quarantine.MsgQuarantineOptOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	k.SetQuarantineOptOut(ctx, toAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgQuarantineOptOutResponse{}, nil
}

func (k Keeper) QuarantineAccept(goCtx context.Context, msg *quarantine.MsgQuarantineAccept) (*quarantine.MsgQuarantineAcceptResponse, error) {
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
		qHolderAddr := k.GetQuarantinedFundsHolder()
		if len(qHolderAddr) == 0 {
			return nil, sdkerrors.ErrUnknownAddress.Wrapf("no quarantine holder account defined")
		}

		if err = k.bankKeeper.SendCoinsBypassQuarantine(ctx, qHolderAddr, toAddr, funds.Coins); err != nil {
			return nil, err
		}
	}

	k.SetQuarantineRecordAccepted(ctx, toAddr, fromAddr)

	if msg.Permanent {
		k.SetQuarantineAutoResponse(ctx, toAddr, fromAddr, quarantine.QUARANTINE_AUTO_RESPONSE_ACCEPT)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgQuarantineAcceptResponse{}, nil
}

func (k Keeper) QuarantineDecline(goCtx context.Context, msg *quarantine.MsgQuarantineDecline) (*quarantine.MsgQuarantineDeclineResponse, error) {
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
		k.SetQuarantineAutoResponse(ctx, toAddr, fromAddr, quarantine.QUARANTINE_AUTO_RESPONSE_DECLINE)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgQuarantineDeclineResponse{}, nil
}

func (k Keeper) UpdateQuarantineAutoResponses(goCtx context.Context, msg *quarantine.MsgUpdateQuarantineAutoResponses) (*quarantine.MsgUpdateQuarantineAutoResponsesResponse, error) {
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
			sdk.NewAttribute(sdk.AttributeKeyModule, quarantine.AttributeValueCategory),
		),
	)

	return &quarantine.MsgUpdateQuarantineAutoResponsesResponse{}, nil
}
