package quarantine

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgOptIn{}

// TODO[1046]: Implement the LegacyMsg interface for MsgOptIn? GetSignBytes() []byte, Route() string, Type() string

// NewMsgOptIn creates a new msg to opt in to account quarantine.
func NewMsgOptIn(toAddr sdk.AccAddress) *MsgOptIn {
	return &MsgOptIn{
		ToAddress: toAddr.String(),
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgOptIn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgOptIn) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgOptOut{}

// TODO[1046]: Implement the LegacyMsg interface for MsgOptOut? GetSignBytes() []byte, Route() string, Type() string

// NewMsgOptOut creates a new msg to opt out of account quarantine.
func NewMsgOptOut(toAddr sdk.AccAddress) *MsgOptOut {
	return &MsgOptOut{
		ToAddress: toAddr.String(),
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgOptOut) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgOptOut) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgAccept{}

// TODO[1046]: Implement the LegacyMsg interface for MsgAccept? GetSignBytes() []byte, Route() string, Type() string

// NewMsgAccept creates a new msg to accept quarantined funds.
func NewMsgAccept(toAddr, fromAddr sdk.AccAddress, permanent bool) *MsgAccept {
	return &MsgAccept{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Permanent:   permanent,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgAccept) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgAccept) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgDecline{}

// TODO[1046]: Implement the LegacyMsg interface for MsgDecline? GetSignBytes() []byte, Route() string, Type() string

// NewMsgDecline creates a new msg to decline quarantined funds.
func NewMsgDecline(toAddr, fromAddr sdk.AccAddress, permanent bool) *MsgDecline {
	return &MsgDecline{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Permanent:   permanent,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgDecline) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgDecline) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgUpdateAutoResponses{}

// TODO[1046]: Implement the LegacyMsg interface for MsgUpdateAutoResponses? GetSignBytes() []byte, Route() string, Type() string

// NewMsgUpdateAutoResponses creates a new msg to update quarantined auto-responses.
func NewMsgUpdateAutoResponses(toAddr sdk.AccAddress, updates []*AutoResponseUpdate) *MsgUpdateAutoResponses {
	return &MsgUpdateAutoResponses{
		ToAddress: toAddr.String(),
		Updates:   updates,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgUpdateAutoResponses) ValidateBasic() error {
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
func (msg MsgUpdateAutoResponses) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}
