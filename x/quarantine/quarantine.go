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

// AcceptFromAddrs moves the provided addrs from the unaccepted slice to the accepted slice.
// If none of the provided addresses are in this record's unaccepted slice, this does nothing.
func (r *QuarantineRecord) AcceptFromAddrs(addrs []sdk.AccAddress) {
	leftoverAddrs := make([]sdk.AccAddress, 0, len(r.UnacceptedFromAddresses))
	for _, existing := range r.UnacceptedFromAddresses {
		if containsAddress(addrs, existing) {
			r.AcceptedFromAddresses = append(r.AcceptedFromAddresses, existing)
		} else {
			leftoverAddrs = append(leftoverAddrs, existing)
		}
	}
	r.UnacceptedFromAddresses = leftoverAddrs
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

// RemoveSuffixes removes the provided suffixes from this.
// Any provided suffixes that are not in this are ignored.
func (s *QuarantineRecordSuffixIndex) RemoveSuffixes(suffixes ...[]byte) {
	firstRem := -1
	for i := 0; i < len(s.RecordSuffixes); i++ {
		if containsSuffix(suffixes, s.RecordSuffixes[i]) {
			firstRem = i
			break
		}
	}
	if firstRem != -1 {
		recordSuffixes := make([][]byte, firstRem, len(s.RecordSuffixes)-1)
		copy(recordSuffixes, s.RecordSuffixes[:firstRem])
		for i := firstRem + 1; i < len(s.RecordSuffixes); i++ {
			if !containsSuffix(suffixes, s.RecordSuffixes[i]) {
				recordSuffixes = append(recordSuffixes, s.RecordSuffixes[i])
			}
		}
		s.RecordSuffixes = recordSuffixes
	}

}

// Simplify updates the suffixes in this so that they are ordered and there aren't any duplicates.
func (s *QuarantineRecordSuffixIndex) Simplify() {
	if len(s.RecordSuffixes) > 1 {
		// Sort the suffixes first, so that deduplication is simpler.
		sort.Slice(s.RecordSuffixes, func(i, j int) bool {
			return bytes.Compare(s.RecordSuffixes[i], s.RecordSuffixes[j]) < 0
		})
		// Do as little work as possible for deduplication.
		// It's assumed that the slice has few duplicates, if any.
		// This is a little extra complex so that the slice isn't just
		// copied every time there aren't any duplicates.
		//
		// First, start going through looking for a dupe.
		// If one is found, record its index and stop.
		// Then, if a first dupe was found, copy the slice up to the entry before it,
		// and iterate over the rest of the suffixes, this time copying the non-duplicates.
		firstDup := -1
		for i := 1; i < len(s.RecordSuffixes); i++ {
			if bytes.Equal(s.RecordSuffixes[i-1], s.RecordSuffixes[i]) {
				firstDup = i
				break
			}
		}
		if firstDup != -1 {
			deduped := make([][]byte, firstDup, len(s.RecordSuffixes)-1)
			copy(deduped, s.RecordSuffixes[:firstDup])
			for i := firstDup + 1; i < len(s.RecordSuffixes); i++ {
				if !bytes.Equal(s.RecordSuffixes[i-1], s.RecordSuffixes[i]) {
					deduped = append(deduped, s.RecordSuffixes[i])
				}
			}
			s.RecordSuffixes = deduped
		}
	}

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
