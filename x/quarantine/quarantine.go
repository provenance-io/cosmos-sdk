package quarantine

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/quarantine/errors"
)

// NewQuarantinedFunds creates a new quarantined funds object.
func NewQuarantinedFunds(toAddr, fromAddr sdk.AccAddress, coins sdk.Coins, declined bool) *QuarantinedFunds {
	return &QuarantinedFunds{
		ToAddress:   toAddr.String(),
		FromAddress: fromAddr.String(),
		Coins:       coins,
		Declined:    declined,
	}
}

// AsQuarantineRecord creates a new QuarantineRecord using the fields in this.
func (f QuarantinedFunds) AsQuarantineRecord() *QuarantineRecord {
	return NewQuarantineRecord(f.Coins, f.Declined)
}

// Validate does simple stateless validation of these quarantined funds.
func (f QuarantinedFunds) Validate() error {
	if _, err := sdk.AccAddressFromBech32(f.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(f.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %v", err)
	}
	if err := f.Coins.Validate(); err != nil {
		return err
	}
	return nil
}

// NewAutoResponseEntry creates a new quarantined auto-response entry.
func NewAutoResponseEntry(toAddr, fromAddr sdk.AccAddress, response AutoResponse) *AutoResponseEntry {
	return &AutoResponseEntry{
		ToAddress:   toAddr.String(),
		FromAddress: fromAddr.String(),
		Response:    response,
	}
}

// Validate does simple stateless validation of these quarantined funds.
func (e AutoResponseEntry) Validate() error {
	if _, err := sdk.AccAddressFromBech32(e.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(e.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %v", err)
	}
	if !e.Response.IsValid() {
		return errors.ErrInvalidValue.Wrapf("unknown auto-response value: %d", e.Response)
	}
	return nil
}

// NewAutoResponseUpdate creates a new quarantine auto-response update.
func NewAutoResponseUpdate(fromAddr sdk.AccAddress, response AutoResponse) *AutoResponseUpdate {
	return &AutoResponseUpdate{
		FromAddress: fromAddr.String(),
		Response:    response,
	}
}

// Validate does simple stateless validation of this update.
func (u AutoResponseUpdate) Validate() error {
	if _, err := sdk.AccAddressFromBech32(u.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}
	if !u.Response.IsValid() {
		return errors.ErrInvalidValue.Wrapf("unknown auto-response value: %d", u.Response)
	}
	return nil
}

const (
	// NoAutoB is a byte with value 0 (corresponding to AUTO_RESPONSE_UNSPECIFIED).
	NoAutoB = byte(0x00)
	// AutoAcceptB is a byte with value 1 (corresponding to AUTO_RESPONSE_ACCEPT).
	AutoAcceptB = byte(0x01)
	// AutoDeclineB is a byte with value 2 (corresponding to AUTO_RESPONSE_DECLINE).
	AutoDeclineB = byte(0x02)
)

// ToAutoB converts a AutoResponse into the byte that will represent it.
func ToAutoB(r AutoResponse) byte {
	switch r {
	case AUTO_RESPONSE_ACCEPT:
		return AutoAcceptB
	case AUTO_RESPONSE_DECLINE:
		return AutoDeclineB
	default:
		return NoAutoB
	}
}

// ToAutoResponse returns the AutoResponse represented by the provided byte slice.
func ToAutoResponse(bz []byte) AutoResponse {
	if len(bz) == 1 {
		switch bz[0] {
		case AutoAcceptB:
			return AUTO_RESPONSE_ACCEPT
		case AutoDeclineB:
			return AUTO_RESPONSE_DECLINE
		}
	}
	return AUTO_RESPONSE_UNSPECIFIED
}

// IsValid returns true if this is a known response value
func (r AutoResponse) IsValid() bool {
	_, found := AutoResponse_name[int32(r)]
	return found
}

// IsAccept returns true if this is an auto-accept response.
func (r AutoResponse) IsAccept() bool {
	return r == AUTO_RESPONSE_ACCEPT
}

// IsDecline returns true if this is an auto-decline response.
func (r AutoResponse) IsDecline() bool {
	return r == AUTO_RESPONSE_DECLINE
}

// NewQuarantineRecord creates a new quarantine record object.
func NewQuarantineRecord(coins sdk.Coins, declined bool) *QuarantineRecord {
	return &QuarantineRecord{
		Coins:    coins,
		Declined: declined,
	}
}

// Validate does simple stateless validation of these quarantined funds.
func (r QuarantineRecord) Validate() error {
	return r.Coins.Validate()
}

// IsZero returns true if this does not have any coins.
func (r QuarantineRecord) IsZero() bool {
	return r.Coins.IsZero()
}

// Add adds coins to this.
func (r *QuarantineRecord) Add(coins ...sdk.Coin) {
	r.Coins.Add(coins...)
}

// AsQuarantinedFunds creates a new QuarantinedFunds using fields in this and the provided addresses.
func (r QuarantineRecord) AsQuarantinedFunds(toAddr, fromAddr sdk.AccAddress) *QuarantinedFunds {
	return NewQuarantinedFunds(toAddr, fromAddr, r.Coins, r.Declined)
}
