package quarantine

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgQuarantineOptIn{}

// TODO[1046]: Implement the LegacyMsg interface for MsgQuarantineOptIn? GetSignBytes() []byte, Route() string, Type() string

// NewMsgQuarantineOptIn creates a new msg to opt in to account quarantine.
func NewMsgQuarantineOptIn(toAddr sdk.AccAddress) *MsgQuarantineOptIn {
	return &MsgQuarantineOptIn{
		ToAddress: toAddr.String(),
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgQuarantineOptIn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgQuarantineOptIn) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgQuarantineOptOut{}

// TODO[1046]: Implement the LegacyMsg interface for MsgQuarantineOptOut? GetSignBytes() []byte, Route() string, Type() string

// NewMsgQuarantineOptOut creates a new msg to opt out of account quarantine.
func NewMsgQuarantineOptOut(toAddr sdk.AccAddress) *MsgQuarantineOptOut {
	return &MsgQuarantineOptOut{
		ToAddress: toAddr.String(),
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgQuarantineOptOut) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgQuarantineOptOut) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgQuarantineAccept{}

// TODO[1046]: Implement the LegacyMsg interface for MsgQuarantineAccept? GetSignBytes() []byte, Route() string, Type() string

// NewMsgQuarantineAccept creates a new msg to accept quarantined funds.
func NewMsgQuarantineAccept(toAddr, fromAddr sdk.AccAddress, permanent bool) *MsgQuarantineAccept {
	return &MsgQuarantineAccept{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Permanent:   permanent,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgQuarantineAccept) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgQuarantineAccept) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgQuarantineDecline{}

// TODO[1046]: Implement the LegacyMsg interface for MsgQuarantineDecline? GetSignBytes() []byte, Route() string, Type() string

// NewMsgQuarantineDecline creates a new msg to decline quarantined funds.
func NewMsgQuarantineDecline(toAddr, fromAddr sdk.AccAddress, permanent bool) *MsgQuarantineDecline {
	return &MsgQuarantineDecline{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Permanent:   permanent,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgQuarantineDecline) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgQuarantineDecline) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgUpdateQuarantineAutoResponses{}

// TODO[1046]: Implement the LegacyMsg interface for MsgUpdateQuarantineAutoResponses? GetSignBytes() []byte, Route() string, Type() string

// NewMsgUpdateQuarantineAutoResponses creates a new msg to update quarantined auto-responses.
func NewMsgUpdateQuarantineAutoResponses(toAddr sdk.AccAddress, updates []*QuarantineAutoResponseUpdate) *MsgUpdateQuarantineAutoResponses {
	return &MsgUpdateQuarantineAutoResponses{
		ToAddress: toAddr.String(),
		Updates:   updates,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgUpdateQuarantineAutoResponses) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	for i, update := range msg.Updates {
		if err := update.Validate(); err != nil {
			return errors.Wrapf(err, "invalid update %d", i)
		}
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgUpdateQuarantineAutoResponses) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}
