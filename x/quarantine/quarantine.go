package quarantine

import (
	"bytes"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/quarantine/errors"
)

// NewQuarantinedFunds creates a new quarantined funds object.
func NewQuarantinedFunds(toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress, coins sdk.Coins, declined bool) *QuarantinedFunds {
	rv := &QuarantinedFunds{
		ToAddress:               toAddr.String(),
		UnacceptedFromAddresses: make([]string, len(fromAddrs)),
		Coins:                   coins,
		Declined:                declined,
	}
	for i, addr := range fromAddrs {
		rv.UnacceptedFromAddresses[i] = addr.String()
	}
	return rv
}

// Validate does simple stateless validation of these quarantined funds.
func (f QuarantinedFunds) Validate() error {
	if _, err := sdk.AccAddressFromBech32(f.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}
	if len(f.UnacceptedFromAddresses) == 0 {
		return errors.ErrInvalidValue.Wrap("at least one unaccepted from address is required")
	}
	seen := make(map[string]struct{})
	for i, addr := range f.UnacceptedFromAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid unaccepted from address[%d]: %v", i, err)
		}
		if _, found := seen[addr]; found {
			return errors.ErrInvalidValue.Wrapf("duplicate unaccepted from address: %q", addr)
		}
		seen[addr] = struct{}{}
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
func NewQuarantineRecord(unacceptedFromAddrs []string, coins sdk.Coins, declined bool) *QuarantineRecord {
	rv := &QuarantineRecord{
		UnacceptedFromAddresses: nil,
		AcceptedFromAddresses:   nil,
		Coins:                   coins,
		Declined:                declined,
	}
	if len(unacceptedFromAddrs) > 0 {
		rv.UnacceptedFromAddresses = make([]sdk.AccAddress, len(unacceptedFromAddrs))
		for i, addr := range unacceptedFromAddrs {
			rv.UnacceptedFromAddresses[i] = sdk.MustAccAddressFromBech32(addr)
		}
	}
	return rv
}

// Validate does simple stateless validation of these quarantined funds.
func (r QuarantineRecord) Validate() error {
	if len(r.UnacceptedFromAddresses) == 0 {
		return errors.ErrInvalidValue.Wrap("at least one unaccepted from address is required")
	}
	return r.Coins.Validate()
}

// IsZero returns true if this does not have any coins.
func (r QuarantineRecord) IsZero() bool {
	return r.Coins.IsZero()
}

// AddCoins adds coins to this.
func (r *QuarantineRecord) AddCoins(coins ...sdk.Coin) {
	r.Coins = r.Coins.Add(coins...)
}

// IsFullyAccepted returns true if this record has been accepted for all from addresses involved.
func (r QuarantineRecord) IsFullyAccepted() bool {
	return len(r.UnacceptedFromAddresses) == 0
}

// AcceptFrom moves the provided addrs from the unaccepted slice to the accepted slice.
// If none of the provided addrs are in this record's unaccepted slice, this does nothing.
// Returns true if one or more addresses were moved to accepted.
func (r *QuarantineRecord) AcceptFrom(addrs []sdk.AccAddress) bool {
	rv := false
	leftoverAddrs := make([]sdk.AccAddress, 0, len(r.UnacceptedFromAddresses))
	for _, existing := range r.UnacceptedFromAddresses {
		if containsAddress(addrs, existing) {
			rv = true
			r.AcceptedFromAddresses = append(r.AcceptedFromAddresses, existing)
		} else {
			leftoverAddrs = append(leftoverAddrs, existing)
		}
	}
	if len(leftoverAddrs) == 0 {
		r.UnacceptedFromAddresses = nil
	} else {
		r.UnacceptedFromAddresses = leftoverAddrs
	}
	return rv
}

func (r *QuarantineRecord) GetAllFromAddrs() []sdk.AccAddress {
	rv := make([]sdk.AccAddress, len(r.UnacceptedFromAddresses)+len(r.AcceptedFromAddresses))
	copy(rv, r.UnacceptedFromAddresses)
	copy(rv[len(r.UnacceptedFromAddresses):], r.AcceptedFromAddresses)
	return rv
}

// containsAddress returns true if the addrToFind is an entry in the addrs.
func containsAddress(addrs []sdk.AccAddress, addrToFind sdk.AccAddress) bool {
	for _, addr := range addrs {
		if addrToFind.Equals(addr) {
			return true
		}
	}
	return false
}

// AsQuarantinedFunds creates a new QuarantinedFunds using fields in this and the provided addresses.
func (r QuarantineRecord) AsQuarantinedFunds(toAddr sdk.AccAddress) *QuarantinedFunds {
	return NewQuarantinedFunds(toAddr, r.UnacceptedFromAddresses, r.Coins, r.Declined)
}

// AddSuffixes adds the provided suffixes to this.
// No attempt is made to deduplicate entries. After using this, you should use Simplify before trying to save it.
func (s *QuarantineRecordSuffixIndex) AddSuffixes(suffixes ...[]byte) {
	s.RecordSuffixes = append(s.RecordSuffixes, suffixes...)
}

// Simplify updates the suffixes in this so that they are ordered and there aren't any duplicates.
func (s *QuarantineRecordSuffixIndex) Simplify(toRemove ...[]byte) {
	switch len(s.RecordSuffixes) {
	case 0:
		// do nothing for now.
	case 1:
		if containsSuffix(toRemove, s.RecordSuffixes[0]) {
			s.RecordSuffixes = nil
		}
	default:
		// Sort the suffixes first, so that deduplication is simpler.
		sort.Slice(s.RecordSuffixes, func(i, j int) bool {
			return bytes.Compare(s.RecordSuffixes[i], s.RecordSuffixes[j]) < 0
		})
		// Do as little work as possible for deduplication.
		// It's assumed that the slice has few duplicates, if any.
		// This is a little extra complex so that the slice isn't just
		// copied every time there aren't any duplicates.

		// func for testing whether an entry is worth keeping.
		isKeeper := func(cur, other []byte) bool {
			return !containsSuffix(toRemove, cur) && !bytes.Equal(cur, other)
		}

		// First, get rid of any non-keepers at the front of the slice since that can be done in-place.
		for len(s.RecordSuffixes) > 0 && !isKeeper(s.RecordSuffixes[0], nil) {
			s.RecordSuffixes = s.RecordSuffixes[1:]
		}

		// Then, look through the rest of the slice looking for one to remove.
		// If one is found, note it and stop.
		firstRem := -1
		for i := 1; i < len(s.RecordSuffixes); i++ {
			if !isKeeper(s.RecordSuffixes[i], s.RecordSuffixes[i-1]) {
				firstRem = i
				break
			}
		}
		// If we found one to remove, we'll then create the new slice that doesn't have
		// the unwanted entries.
		if firstRem != -1 {
			suffixes := make([][]byte, firstRem, len(s.RecordSuffixes)-1)
			copy(suffixes, s.RecordSuffixes[:firstRem])
			for i := firstRem + 1; i < len(s.RecordSuffixes); i++ {
				if isKeeper(s.RecordSuffixes[i], s.RecordSuffixes[i-1]) {
					suffixes = append(suffixes, s.RecordSuffixes[i])
				}
			}
			s.RecordSuffixes = suffixes
		}
	}

	// If there's nothing left, make sure it's nil.
	if len(s.RecordSuffixes) == 0 {
		s.RecordSuffixes = nil
	}
}

// containsSuffix returns true if the suffixToFind is in the suffixes.
func containsSuffix(suffixes [][]byte, suffixToFind []byte) bool {
	for _, suffix := range suffixes {
		if bytes.Equal(suffixToFind, suffix) {
			return true
		}
	}
	return false
}
