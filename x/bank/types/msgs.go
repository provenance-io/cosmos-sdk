package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// bank message types
const (
	TypeMsgSend      = "send"
	TypeMsgMultiSend = "multisend"
)

var _ sdk.Msg = &MsgSend{}

// NewMsgSend - construct a msg to send coins from one account to another.
//
//nolint:interfacer
func NewMsgSend(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgSend {
	return &MsgSend{FromAddress: fromAddr.String(), ToAddress: toAddr.String(), Amount: amount}
}

// Route Implements Msg.
func (msg MsgSend) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSend) Type() string { return TypeMsgSend }

// ValidateBasic Implements Msg.
func (msg MsgSend) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgSend) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

var _ sdk.Msg = &MsgMultiSend{}

// NewMsgMultiSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgMultiSend(in []Input, out []Output) *MsgMultiSend {
	return &MsgMultiSend{Inputs: in, Outputs: out}
}

// Route Implements Msg
func (msg MsgMultiSend) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgMultiSend) Type() string { return TypeMsgMultiSend }

// ValidateBasic Implements Msg.
func (msg MsgMultiSend) ValidateBasic() error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs
	}

	if len(msg.Outputs) == 0 {
		return ErrNoOutputs
	}

	return ValidateInputsOutputs(msg.Inputs, msg.Outputs)
}

// GetSignBytes Implements Msg.
func (msg MsgMultiSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgMultiSend) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(msg.Inputs))
	for i, in := range msg.Inputs {
		inAddr, _ := sdk.AccAddressFromBech32(in.Address)
		addrs[i] = inAddr
	}

	return addrs
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(in.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid input address: %s", err)
	}

	if !in.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, in.Coins.String())
	}

	if !in.Coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, in.Coins.String())
	}

	return nil
}

// NewInput - create a transaction input, used with MsgMultiSend
//
//nolint:interfacer
func NewInput(addr sdk.AccAddress, coins sdk.Coins) Input {
	return Input{
		Address: addr.String(),
		Coins:   coins,
	}
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(out.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid output address: %s", err)
	}

	if !out.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, out.Coins.String())
	}

	if !out.Coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, out.Coins.String())
	}

	return nil
}

// NewOutput - create a transaction output, used with MsgMultiSend
//
//nolint:interfacer
func NewOutput(addr sdk.AccAddress, coins sdk.Coins) Output {
	return Output{
		Address: addr.String(),
		Coins:   coins,
	}
}

// ValidateInputsOutputs validates that each respective input and output is
// valid and that the sum of inputs is equal to the sum of outputs.
func ValidateInputsOutputs(inputs []Input, outputs []Output) error {
	var totalIn, totalOut sdk.Coins

	for _, in := range inputs {
		if err := in.ValidateBasic(); err != nil {
			return err
		}

		totalIn = totalIn.Add(in.Coins...)
	}

	for _, out := range outputs {
		if err := out.ValidateBasic(); err != nil {
			return err
		}

		totalOut = totalOut.Add(out.Coins...)
	}

	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return ErrInputOutputMismatch
	}

	return nil
}

var _ sdk.Msg = &MsgQuarantineOptIn{}

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
		if err := update.ValidateBasic(); err != nil {
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

// NewQuarantineAutoResponseUpdate creates a new quarantine auto-response update.
func NewQuarantineAutoResponseUpdate(fromAddr sdk.AccAddress, response QuarantineAutoResponse) *QuarantineAutoResponseUpdate {
	return &QuarantineAutoResponseUpdate{
		FromAddress: fromAddr.String(),
		Response:    response,
	}
}

// ValidateBasic does simple stateless validation of this update.
func (u QuarantineAutoResponseUpdate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(u.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}
	if _, found := QuarantineAutoResponse_name[int32(u.Response)]; !found {
		return ErrInvalidValue.Wrapf("unknown response value: %d", u.Response)
	}
	return nil
}
