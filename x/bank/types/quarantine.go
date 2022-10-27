package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// NoAutoB is a byte with value 0 (corresponding to QUARANTINE_AUTO_RESPONSE_UNSPECIFIED).
	NoAutoB = byte(0x00)
	// AutoAcceptB is a byte with value 1 (corresponding to QUARANTINE_AUTO_RESPONSE_ACCEPT).
	AutoAcceptB = byte(0x01)
	// AutoDeclineB is a byte with value 2 (corresponding to QUARANTINE_AUTO_RESPONSE_DECLINE).
	AutoDeclineB = byte(0x02)
)

// ToAutoB converts a QuarantineAutoResponse into the byte that will represent it.
func ToAutoB(r QuarantineAutoResponse) byte {
	switch r {
	case QUARANTINE_AUTO_RESPONSE_ACCEPT:
		return AutoAcceptB
	case QUARANTINE_AUTO_RESPONSE_DECLINE:
		return AutoDeclineB
	default:
		return NoAutoB
	}
}

// ToQuarantineAutoResponse returns the QuarantineAutoResponse represented by the provided byte slice.
func ToQuarantineAutoResponse(bz []byte) QuarantineAutoResponse {
	if len(bz) == 1 {
		switch bz[0] {
		case AutoAcceptB:
			return QUARANTINE_AUTO_RESPONSE_ACCEPT
		case AutoDeclineB:
			return QUARANTINE_AUTO_RESPONSE_DECLINE
		}
	}
	return QUARANTINE_AUTO_RESPONSE_UNSPECIFIED
}

// IsAutoAcceptB returns true if the provided byte slice has exactly one byte, and it is equal to AutoAccept.
func IsAutoAcceptB(bz []byte) bool {
	return len(bz) == 1 && bz[0] == AutoAcceptB
}

// IsAutoDeclineB returns true if the provided byte slice has exactly one byte, and it is equal to AutoDecline.
func IsAutoDeclineB(bz []byte) bool {
	return len(bz) == 1 && bz[0] == AutoDeclineB
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

// NewQuarantineRecord creates a new quarantine record object.
func NewQuarantineRecord(coins sdk.Coins, declined bool) *QuarantineRecord {
	return &QuarantineRecord{
		Coins:    coins,
		Declined: declined,
	}
}

// ValidateBasic does simple stateless validation of these quarantined funds.
func (r QuarantineRecord) ValidateBasic() error {
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
	return &QuarantinedFunds{
		ToAddress:   toAddr.String(),
		FromAddress: fromAddr.String(),
		Coins:       r.Coins,
		Declined:    r.Declined,
	}
}

// NewQuarantinedFunds creates a new quarantined funds object.
func NewQuarantinedFunds(toAddr, fromAddr sdk.AccAddress, coins sdk.Coins, declined bool) *QuarantinedFunds {
	return &QuarantinedFunds{
		ToAddress:   toAddr.String(),
		FromAddress: fromAddr.String(),
		Coins:       coins,
		Declined:    declined,
	}
}
