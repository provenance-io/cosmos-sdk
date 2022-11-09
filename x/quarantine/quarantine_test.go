package quarantine

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
)

// makeTestAddr makes an AccAddress for use in tests.
// The base and index are turned into a string, then hashed.
// Then the first byte is changed to the index and the next bytes are the base.
// This can help identify it in failed test output.
func makeTestAddr(base string, index uint8) sdk.AccAddress {
	toHash := fmt.Sprintf("%s test address %d", base, index)
	rv := sdk.AccAddress(crypto.AddressHash([]byte(toHash)))
	rv[0] = index
	copy(rv[1:], base)
	return rv
}

type coinMaker func() sdk.Coins

var (
	coinMakerOK    coinMaker = func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("okcoin", 100)) }
	coinMakerMulti coinMaker = func() sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin("multicoina", 33), sdk.NewInt64Coin("multicoinb", 67))
	}
	coinMakerEmpty coinMaker = func() sdk.Coins { return sdk.Coins{} }
	coinMakerNil   coinMaker = func() sdk.Coins { return nil }
	coinMakerBad   coinMaker = func() sdk.Coins { return sdk.Coins{sdk.Coin{Denom: "badcoin", Amount: sdk.NewInt(-1)}} }
)

// assertErrorContents asserts that, if contains is empty, there's no error.
// Otherwise, asserts that there is an error, and that it contains each of the provided strings.
func assertErrorContents(t *testing.T, theError error, contains []string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(contains) == 0 {
		return assert.NoError(t, theError, msgAndArgs)
	}
	rv := assert.Error(t, theError, msgAndArgs...)
	if rv {
		for _, expInErr := range contains {
			rv = assert.ErrorContains(t, theError, expInErr, msgAndArgs...) && rv
		}
	}
	return rv
}

// makeCopyOfCoins creates a copy of the provided Coins and returns it.
func makeCopyOfCoins(orig sdk.Coins) sdk.Coins {
	if orig == nil {
		return nil
	}
	rv := make(sdk.Coins, len(orig))
	for i, coin := range orig {
		rv[i] = sdk.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.AddRaw(0),
		}
	}
	return rv
}

// makeCopyOfQuarantinedFunds creates a copy of the provided QuarantinedFunds and returns it.
func makeCopyOfQuarantinedFunds(orig *QuarantinedFunds) *QuarantinedFunds {
	return &QuarantinedFunds{
		ToAddress:               orig.ToAddress,
		UnacceptedFromAddresses: makeCopyOfStringSlice(orig.UnacceptedFromAddresses),
		Coins:                   makeCopyOfCoins(orig.Coins),
		Declined:                orig.Declined,
	}
}

// makeCopyOfStringSlice makes a copy of the provided string slice.
func makeCopyOfStringSlice(orig []string) []string {
	if orig == nil {
		return nil
	}
	rv := make([]string, len(orig))
	copy(rv, orig)
	return rv
}

// makeCopyOfQuarantineRecord creates a copy of the provided QuarantineRecord and returns it.
func makeCopyOfQuarantineRecord(orig *QuarantineRecord) *QuarantineRecord {
	return &QuarantineRecord{
		UnacceptedFromAddresses: makeCopyOfAccAddresses(orig.UnacceptedFromAddresses),
		AcceptedFromAddresses:   makeCopyOfAccAddresses(orig.AcceptedFromAddresses),
		Coins:                   makeCopyOfCoins(orig.Coins),
		Declined:                orig.Declined,
	}
}

// makeCopyOfAccAddresses makes a copy of the provided slice of acc addresses, copying each address too.
func makeCopyOfAccAddresses(orig []sdk.AccAddress) []sdk.AccAddress {
	if orig == nil {
		return nil
	}
	rv := make([]sdk.AccAddress, len(orig))
	for i, addr := range orig {
		rv[i] = make(sdk.AccAddress, len(addr))
		copy(rv[i], addr)
	}
	return rv
}

func TestNewQuarantinedFunds(t *testing.T) {
	testAddr0 := makeTestAddr("nqf", 0)
	testAddr1 := makeTestAddr("nqf", 1)

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []sdk.AccAddress
		Coins     sdk.Coins
		declined  bool
		expected  *QuarantinedFunds
	}{
		{
			name:      "control",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 88)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 88)),
				Declined:                false,
			},
		},
		{
			name:      "declined true",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 87)),
			declined:  true,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 87)),
				Declined:                true,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 86)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 86)),
				Declined:                false,
			},
		},
		{
			name:      "nil fromAddrs",
			toAddr:    testAddr0,
			fromAddrs: nil,
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
				Declined:                false,
			},
		},
		{
			name:      "empty fromAddrs",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
				Declined:                false,
			},
		},
		{
			name:      "empty coins",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.Coins{},
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.Coins{},
				Declined:                false,
			},
		},
		{
			name:      "nil coins",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     nil,
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   nil,
				Declined:                false,
			},
		},
		{
			name:      "invalid coins",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.Coins{sdk.Coin{Denom: "bad", Amount: sdk.NewInt(-1)}},
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.Coins{sdk.Coin{Denom: "bad", Amount: sdk.NewInt(-1)}},
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewQuarantinedFunds(tc.toAddr, tc.fromAddrs, tc.Coins, tc.declined)
			assert.Equal(t, tc.expected, actual, "NewQuarantinedFunds")
		})
	}
}

func TestQuarantinedFunds_Validate(t *testing.T) {
	testAddr0 := makeTestAddr("qfv", 0).String()
	testAddr1 := makeTestAddr("qfv", 1).String()
	testAddr2 := makeTestAddr("qfv", 2).String()

	tests := []struct {
		name          string
		qf            *QuarantinedFunds
		expectedInErr []string
	}{
		{
			name: "control",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "declined true",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expectedInErr: nil,
		},
		{
			name: "bad to address",
			qf: &QuarantinedFunds{
				ToAddress:               "notgonnawork",
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "empty to address",
			qf: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "bad from address",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{"alsonotgood"},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "empty from address",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{""},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "nil from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required", "invalid value"},
		},
		{
			name: "empty from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required", "invalid value"},
		},
		{
			name: "two from addresses both good",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1, testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "two same from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr2, testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"duplicate unaccepted from address", testAddr2},
		},
		{
			name: "three from addresses same first last",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1, testAddr2, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"duplicate unaccepted from address", testAddr1},
		},
		{
			name: "two from addresses first bad",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{"this is not an address", testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "two from addresses last bad",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1, "this is also bad"},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[1]"},
		},
		{
			name: "empty coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "nil coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerNil(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "bad coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
			expectedInErr: []string{coinMakerBad().String(), "amount is not positive"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qfOrig := makeCopyOfQuarantinedFunds(tc.qf)
			err := tc.qf.Validate()
			assertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, qfOrig, tc.qf, "QuarantinedFunds before and after")
		})
	}
}

func TestNewAutoResponseEntry(t *testing.T) {
	testAddr0 := makeTestAddr("nare", 0)
	testAddr1 := makeTestAddr("nare", 1)

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		resp     AutoResponse
		expected *AutoResponseEntry
	}{
		{
			name:     "accept",
			toAddr:   testAddr0,
			fromAddr: testAddr1,
			resp:     AUTO_RESPONSE_ACCEPT,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr0.String(),
				FromAddress: testAddr1.String(),
				Response:    AUTO_RESPONSE_ACCEPT,
			},
		},
		{
			name:     "decline",
			toAddr:   testAddr0,
			fromAddr: testAddr1,
			resp:     AUTO_RESPONSE_DECLINE,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr0.String(),
				FromAddress: testAddr1.String(),
				Response:    AUTO_RESPONSE_DECLINE,
			},
		},
		{
			name:     "unspecified",
			toAddr:   testAddr0,
			fromAddr: testAddr1,
			resp:     AUTO_RESPONSE_UNSPECIFIED,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr0.String(),
				FromAddress: testAddr1.String(),
				Response:    AUTO_RESPONSE_UNSPECIFIED,
			},
		},
		{
			name:     "nil to address",
			toAddr:   nil,
			fromAddr: testAddr1,
			resp:     AUTO_RESPONSE_ACCEPT,
			expected: &AutoResponseEntry{
				ToAddress:   "",
				FromAddress: testAddr1.String(),
				Response:    AUTO_RESPONSE_ACCEPT,
			},
		},
		{
			name:     "nil from address",
			toAddr:   testAddr0,
			fromAddr: nil,
			resp:     AUTO_RESPONSE_DECLINE,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr0.String(),
				FromAddress: "",
				Response:    AUTO_RESPONSE_DECLINE,
			},
		},
		{
			name:     "weird response",
			toAddr:   testAddr1,
			fromAddr: testAddr0,
			resp:     -3,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr1.String(),
				FromAddress: testAddr0.String(),
				Response:    -3,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewAutoResponseEntry(tc.toAddr, tc.fromAddr, tc.resp)
			assert.Equal(t, tc.expected, actual, "NewAutoResponseEntry")
		})
	}
}

func TestAutoResponseEntry_Validate(t *testing.T) {
	testAddr0 := makeTestAddr("arev", 0).String()
	testAddr1 := makeTestAddr("arev", 1).String()

	tests := []struct {
		name          string
		toAddr        string
		fromAddr      string
		resp          AutoResponse
		qf            QuarantinedFunds
		expectedInErr []string
	}{
		{
			name:          "accept",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: nil,
		},
		{
			name:          "decline",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: nil,
		},
		{
			name:          "unspecified",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: nil,
		},
		{
			name:          "bad to address",
			toAddr:        "notgonnawork",
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad from address",
			toAddr:        testAddr0,
			fromAddr:      "alsonotgood",
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "empty from address",
			toAddr:        testAddr0,
			fromAddr:      "",
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "negative response",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          -1,
			expectedInErr: []string{"unknown auto-response value", "-1"},
		},
		{
			name:          "response too large",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          3,
			expectedInErr: []string{"unknown auto-response value", "3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entryOrig := AutoResponseEntry{
				ToAddress:   tc.toAddr,
				FromAddress: tc.fromAddr,
				Response:    tc.resp,
			}
			entry := AutoResponseEntry{
				ToAddress:   tc.toAddr,
				FromAddress: tc.fromAddr,
				Response:    tc.resp,
			}
			err := entry.Validate()
			assertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, entryOrig, entry, "AutoResponseEntry before and after")
		})
	}
}

func TestAutoResponseUpdate_Validate(t *testing.T) {
	testAddr0 := makeTestAddr("arev", 0).String()
	testAddr1 := makeTestAddr("arev", 1).String()

	tests := []struct {
		name          string
		fromAddr      string
		resp          AutoResponse
		qf            QuarantinedFunds
		expectedInErr []string
	}{
		{
			name:          "accept",
			fromAddr:      testAddr0,
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: nil,
		},
		{
			name:          "decline",
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: nil,
		},
		{
			name:          "unspecified",
			fromAddr:      testAddr0,
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: nil,
		},
		{
			name:          "bad from address",
			fromAddr:      "yupnotgood",
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "empty from address",
			fromAddr:      "",
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "negative response",
			fromAddr:      testAddr1,
			resp:          -1,
			expectedInErr: []string{"unknown auto-response value", "-1"},
		},
		{
			name:          "response too large",
			fromAddr:      testAddr0,
			resp:          3,
			expectedInErr: []string{"unknown auto-response value", "3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updateOrig := AutoResponseUpdate{
				FromAddress: tc.fromAddr,
				Response:    tc.resp,
			}
			update := AutoResponseUpdate{
				FromAddress: tc.fromAddr,
				Response:    tc.resp,
			}
			err := update.Validate()
			assertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, updateOrig, update, "AutoResponseUpdate before and after")
		})
	}
}

func TestAutoBValues(t *testing.T) {
	// If these were the same, it'd be bad.
	assert.NotEqual(t, NoAutoB, AutoAcceptB, "NoAutoB vs AutoAcceptB")
	assert.NotEqual(t, NoAutoB, AutoDeclineB, "NoAutoB vs AutoDeclineB")
	assert.NotEqual(t, AutoAcceptB, AutoDeclineB, "AutoAcceptB vs AutoDeclineB")
}

func TestToAutoB(t *testing.T) {
	tests := []struct {
		name     string
		r        AutoResponse
		expected byte
	}{
		{
			name:     "accept",
			r:        AUTO_RESPONSE_ACCEPT,
			expected: AutoAcceptB,
		},
		{
			name:     "decline",
			r:        AUTO_RESPONSE_DECLINE,
			expected: AutoDeclineB,
		},
		{
			name:     "unspecified",
			r:        AUTO_RESPONSE_UNSPECIFIED,
			expected: NoAutoB,
		},
		{
			name:     "negative",
			r:        -1,
			expected: NoAutoB,
		},
		{
			name:     "too large",
			r:        3,
			expected: NoAutoB,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := ToAutoB(tc.r)
			assert.Equal(t, tc.expected, actual, "ToAutoB(%s)", tc.r)
		})
	}
}

func TestAutoResponseValues(t *testing.T) {
	// If these were the same, it'd be bad.
	assert.NotEqual(t, AUTO_RESPONSE_UNSPECIFIED, AUTO_RESPONSE_ACCEPT, "AUTO_RESPONSE_UNSPECIFIED vs AUTO_RESPONSE_ACCEPT")
	assert.NotEqual(t, AUTO_RESPONSE_UNSPECIFIED, AUTO_RESPONSE_DECLINE, "AUTO_RESPONSE_UNSPECIFIED vs AUTO_RESPONSE_DECLINE")
	assert.NotEqual(t, AUTO_RESPONSE_ACCEPT, AutoDeclineB, "AUTO_RESPONSE_ACCEPT vs AUTO_RESPONSE_DECLINE")
}

func TestToAutoResponse(t *testing.T) {
	tests := []struct {
		name     string
		bz       []byte
		expected AutoResponse
	}{
		{
			name:     "accept",
			bz:       []byte{AutoAcceptB},
			expected: AUTO_RESPONSE_ACCEPT,
		},
		{
			name:     "decline",
			bz:       []byte{AutoDeclineB},
			expected: AUTO_RESPONSE_DECLINE,
		},
		{
			name:     "unspecified",
			bz:       []byte{NoAutoB},
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
		{
			name:     "nil",
			bz:       nil,
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
		{
			name:     "empty",
			bz:       []byte{},
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
		{
			name:     "too long",
			bz:       []byte{AutoAcceptB, AutoAcceptB},
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := ToAutoResponse(tc.bz)
			assert.Equal(t, tc.expected, actual, "ToAutoResponse(%v)", tc.bz)
		})
	}
}

func TestAutoResponse_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		r        AutoResponse
		expected bool
	}{
		{
			name:     "accept",
			r:        AUTO_RESPONSE_ACCEPT,
			expected: true,
		},
		{
			name:     "decline",
			r:        AUTO_RESPONSE_DECLINE,
			expected: true,
		},
		{
			name:     "unspecified",
			r:        AUTO_RESPONSE_UNSPECIFIED,
			expected: true,
		},
		{
			name:     "negative",
			r:        -1,
			expected: false,
		},
		{
			name:     "too large",
			r:        3,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.r
			actual := r.IsValid()
			assert.Equal(t, tc.expected, actual, "%s.IsValid", tc.r)
			assert.Equal(t, tc.r, r, "AutoResponse before and after")
		})
	}
}

func TestAutoResponse_IsAccept(t *testing.T) {
	tests := []struct {
		name     string
		r        AutoResponse
		expected bool
	}{
		{
			name:     "accept",
			r:        AUTO_RESPONSE_ACCEPT,
			expected: true,
		},
		{
			name:     "decline",
			r:        AUTO_RESPONSE_DECLINE,
			expected: false,
		},
		{
			name:     "unspecified",
			r:        AUTO_RESPONSE_UNSPECIFIED,
			expected: false,
		},
		{
			name:     "negative",
			r:        -1,
			expected: false,
		},
		{
			name:     "too large",
			r:        3,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.r
			actual := r.IsAccept()
			assert.Equal(t, tc.expected, actual, "%s.IsAccept", tc.r)
			assert.Equal(t, tc.r, r, "AutoResponse before and after")
		})
	}
}

func TestAutoResponse_IsDecline(t *testing.T) {
	tests := []struct {
		name     string
		r        AutoResponse
		expected bool
	}{
		{
			name:     "accept",
			r:        AUTO_RESPONSE_ACCEPT,
			expected: false,
		},
		{
			name:     "decline",
			r:        AUTO_RESPONSE_DECLINE,
			expected: true,
		},
		{
			name:     "unspecified",
			r:        AUTO_RESPONSE_UNSPECIFIED,
			expected: false,
		},
		{
			name:     "negative",
			r:        -1,
			expected: false,
		},
		{
			name:     "too large",
			r:        3,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.r
			actual := r.IsDecline()
			assert.Equal(t, tc.expected, actual, "%s.IsDecline", tc.r)
			assert.Equal(t, tc.r, r, "AutoResponse before and after")
		})
	}
}

func TestNewQuarantineRecord(t *testing.T) {
	testAddr0 := makeTestAddr("nqr", 0)
	testAddr1 := makeTestAddr("nqr", 1)

	tests := []struct {
		name        string
		uaFromAddrs []string
		coins       sdk.Coins
		declined    bool
		expected    *QuarantineRecord
		expPanic    string
	}{
		{
			name:        "control",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerOK(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name:        "declined",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerOK(),
			declined:    true,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name:        "multi coins",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerMulti(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerMulti(),
				Declined:                false,
			},
		},
		{
			name:        "empty coins",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerEmpty(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
		},
		{
			name:        "nil coins",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerNil(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerNil(),
				Declined:                false,
			},
		},
		{
			name:        "bad coins",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerBad(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
		},
		{
			name:        "bad unaccepted addr panics",
			uaFromAddrs: []string{"I'm a bad address"},
			coins:       coinMakerOK(),
			declined:    false,
			expPanic:    "decoding bech32 failed: string not all lowercase or all uppercase",
		},
		{
			name:        "empty unaccepted addr string panics",
			uaFromAddrs: []string{""},
			coins:       coinMakerOK(),
			declined:    false,
			expPanic:    "empty address string is not allowed",
		},
		{
			name:        "bad second unaccepted addr panics",
			uaFromAddrs: []string{testAddr0.String(), "I'm a bad address"},
			coins:       coinMakerOK(),
			declined:    false,
			expPanic:    "decoding bech32 failed: string not all lowercase or all uppercase",
		},
		{
			name:        "two unaccepted addresses",
			uaFromAddrs: []string{testAddr0.String(), testAddr1.String()},
			coins:       coinMakerOK(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name:        "empty unaccepted addresses",
			uaFromAddrs: []string{},
			coins:       coinMakerOK(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name:        "nil unaccepted addresses",
			uaFromAddrs: nil,
			coins:       coinMakerOK(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *QuarantineRecord
			testFunc := func() {
				actual = NewQuarantineRecord(tc.uaFromAddrs, tc.coins, tc.declined)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "NewQuarantineRecord") {
					assert.Equal(t, tc.expected, actual, "NewQuarantineRecord")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "NewQuarantineRecord")
			}
		})
	}
}

func TestQuarantineRecord_Validate(t *testing.T) {
	testAddr0 := makeTestAddr("qrv", 0)
	testAddr1 := makeTestAddr("qrv", 1)
	testAddr2 := makeTestAddr("qrv", 2)

	tests := []struct {
		name          string
		qr            *QuarantineRecord
		expectedInErr []string
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expectedInErr: nil,
		},
		{
			name: "no accepted addresses is ok",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "nil accepted addresses is ok",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "multi coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerMulti(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "empty coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "nil coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerNil(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "bad coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
			expectedInErr: []string{coinMakerBad().String(), "amount is not positive"},
		},
		{
			name: "nil unaccepted addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required"},
		},
		{
			name: "empty unaccepted addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required"},
		},
		{
			name: "two unaccepted addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "three unaccepted addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1, testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qrOrig := makeCopyOfQuarantineRecord(tc.qr)
			err := tc.qr.Validate()
			assertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, qrOrig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecord_IsZero(t *testing.T) {
	testAddr0 := makeTestAddr("qriz", 0)

	tests := []struct {
		name     string
		qr       *QuarantineRecord
		expected bool
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expected: false,
		},
		{
			name: "multi coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerMulti(),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "empty coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "nil coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerNil(),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "bad coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qrOrig := makeCopyOfQuarantineRecord(tc.qr)
			actual := tc.qr.IsZero()
			assert.Equal(t, tc.expected, actual, "IsZero()")
			assert.Equal(t, qrOrig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecord_AddCoins(t *testing.T) {
	testAddr0 := makeTestAddr("qrac", 0)
	testAddr1 := makeTestAddr("qrac", 1)
	testAddr2 := makeTestAddr("qrac", 2)
	testAddr3 := makeTestAddr("qrac", 3)

	keyEmpty := "empty"
	keyNil := "nil"
	key0Acorn := "0acorn"
	key50Acorn := "50acorn"
	key32Almond := "32almond"
	key8acorn9Almond := "8acorn,9almond"
	coinMakerMap := map[string]coinMaker{
		keyEmpty:         coinMakerEmpty,
		keyNil:           coinMakerNil,
		key0Acorn:        func() sdk.Coins { return sdk.Coins{sdk.NewInt64Coin("acorn", 0)} },
		key50Acorn:       func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)) },
		key32Almond:      func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("almond", 32)) },
		key8acorn9Almond: func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)) },
	}

	tests := []struct {
		qrCoinKey  string
		addCoinKey string
		expected   sdk.Coins
	}{
		// empty
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: keyEmpty,
			expected:   nil,
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: keyNil,
			expected:   nil,
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: key0Acorn,
			expected:   nil,
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// nil
		{
			qrCoinKey:  keyNil,
			addCoinKey: keyEmpty,
			expected:   nil,
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: keyNil,
			expected:   nil,
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: key0Acorn,
			expected:   nil,
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// 0acorn
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: keyEmpty,
			expected:   nil,
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: keyNil,
			expected:   nil,
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: key0Acorn,
			expected:   nil,
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// 50acorn
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: keyEmpty,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: keyNil,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: key0Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 100)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50), sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 58), sdk.NewInt64Coin("almond", 9)),
		},

		// 32almond
		{
			qrCoinKey:  key32Almond,
			addCoinKey: keyEmpty,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: keyNil,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: key0Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50), sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 64)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 41)),
		},

		// 8acorn,9almond
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: keyEmpty,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: keyNil,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: key0Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 58), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 41)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 16), sdk.NewInt64Coin("almond", 18)),
		},
	}

	addressCombos := []struct {
		name       string
		unaccepted []sdk.AccAddress
		accepted   []sdk.AccAddress
	}{
		{
			name:       "no addresses",
			unaccepted: nil,
			accepted:   nil,
		},
		{
			name:       "one unaccepted",
			unaccepted: []sdk.AccAddress{testAddr0},
			accepted:   nil,
		},
		{
			name:       "two unaccepted",
			unaccepted: []sdk.AccAddress{testAddr0, testAddr1},
			accepted:   nil,
		},
		{
			name:       "one accepted",
			unaccepted: nil,
			accepted:   []sdk.AccAddress{testAddr2},
		},
		{
			name:       "two accepted",
			unaccepted: nil,
			accepted:   []sdk.AccAddress{testAddr2, testAddr3},
		},
		{
			name:       "one unaccepted one accepted",
			unaccepted: []sdk.AccAddress{testAddr0},
			accepted:   []sdk.AccAddress{testAddr2},
		},
		{
			name:       "two unaccepted one accepted",
			unaccepted: []sdk.AccAddress{testAddr0, testAddr1},
			accepted:   []sdk.AccAddress{testAddr2},
		},
		{
			name:       "one unaccepted two accepted",
			unaccepted: []sdk.AccAddress{testAddr0},
			accepted:   []sdk.AccAddress{testAddr2, testAddr3},
		},
		{
			name:       "two unaccepted two accepted",
			unaccepted: []sdk.AccAddress{testAddr0, testAddr1},
			accepted:   []sdk.AccAddress{testAddr2, testAddr3},
		},
	}

	for _, tc := range tests {
		for _, ac := range addressCombos {
			for _, declined := range []bool{false, true} {
				name := fmt.Sprintf("%s+%s=%q %t %s", tc.qrCoinKey, tc.addCoinKey, tc.expected.String(), declined, ac.name)
				t.Run(name, func(t *testing.T) {
					expected := QuarantineRecord{
						UnacceptedFromAddresses: makeCopyOfAccAddresses(ac.unaccepted),
						AcceptedFromAddresses:   makeCopyOfAccAddresses(ac.accepted),
						Coins:                   tc.expected,
						Declined:                declined,
					}
					qr := QuarantineRecord{
						UnacceptedFromAddresses: ac.unaccepted,
						AcceptedFromAddresses:   ac.accepted,
						Coins:                   coinMakerMap[tc.qrCoinKey](),
						Declined:                declined,
					}
					addCoinsOrig := coinMakerMap[tc.addCoinKey]()
					addCoins := coinMakerMap[tc.addCoinKey]()
					qr.AddCoins(addCoins...)
					assert.Equal(t, expected, qr, "QuarantineRecord after AddCoins")
					assert.Equal(t, addCoinsOrig, addCoins, "Coins before and after")
				})
			}
		}
	}
}

func TestQuarantineRecord_IsFullyAccepted(t *testing.T) {
	testAddr0 := makeTestAddr("qrifa", 0)
	testAddr1 := makeTestAddr("qrifa", 1)

	tests := []struct {
		name     string
		qr       *QuarantineRecord
		expected bool
	}{
		{
			name: "no addresses at all",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "one unaccepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "declined one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expected: true,
		},
		{
			name: "one accepted one unaccepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "two unaccepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "declined two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := makeCopyOfQuarantineRecord(tc.qr)
			actual := tc.qr.IsFullyAccepted()
			assert.Equal(t, tc.expected, actual, "IsFullyAccepted: %v", tc.qr)
			assert.Equal(t, orig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecord_AcceptFrom(t *testing.T) {
	testAddr0 := makeTestAddr("qraf", 0)
	testAddr1 := makeTestAddr("qraf", 1)
	testAddr2 := makeTestAddr("qraf", 2)
	testAddr3 := makeTestAddr("qraf", 3)
	testAddr4 := makeTestAddr("qraf", 4)

	tests := []struct {
		name  string
		qr    *QuarantineRecord
		addrs []sdk.AccAddress
		exp   bool
		expQr *QuarantineRecord
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "nil addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: nil,
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "empty addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "one addrs only in accepted already",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr1},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "record has nil addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "one address in both",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted two other provided",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr2, testAddr3},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted both provided",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr1},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted both provided opposite order",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr1, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted first provided first with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr3, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted first provided second with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr3, testAddr0, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted first provided third with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr4, testAddr3, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two same unaccepted provided once",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr0, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted second provided first with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr1, testAddr3, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted second provided second with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr3, testAddr1, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted second provided third with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr4, testAddr3, testAddr1},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "one unaccepted provided thrice",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr4},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr0, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr4, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origInput := makeCopyOfAccAddresses(tc.addrs)
			actual := tc.qr.AcceptFrom(tc.addrs)
			assert.Equal(t, tc.exp, actual, "AcceptFrom return value")
			assert.Equal(t, tc.expQr, tc.qr, "QuarantineRecord after AcceptFrom")
			assert.Equal(t, origInput, tc.addrs, "input address slice before and after AcceptFrom")
		})
	}
}

func TestQuarantineRecord_GetAllFromAddrs(t *testing.T) {
	testAddr0 := makeTestAddr("qrgafa", 0)
	testAddr1 := makeTestAddr("qrgafa", 1)
	testAddr2 := makeTestAddr("qrgafa", 2)
	testAddr3 := makeTestAddr("qrgafa", 3)
	testAddr4 := makeTestAddr("qrgafa", 4)

	tests := []struct {
		name string
		qr   *QuarantineRecord
		exp  []sdk.AccAddress
	}{
		{
			name: "nil unaccepted nil accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "nil unaccepted empty accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{},
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "empty unaccepted nil accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   nil,
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "empty unaccepted empty accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{},
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "one unaccepted nil accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
			},
			exp: []sdk.AccAddress{testAddr0},
		},
		{
			name: "two unaccepted nil accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "one unaccepted empty accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{},
			},
			exp: []sdk.AccAddress{testAddr0},
		},
		{
			name: "two unaccepted empty accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "nil unaccepted one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
			},
			exp: []sdk.AccAddress{testAddr0},
		},
		{
			name: "nil unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "empty unaccepted one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
			},
			exp: []sdk.AccAddress{testAddr0},
		},
		{
			name: "empty unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "one unaccepted one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "two unaccepted one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1, testAddr2},
		},
		{
			name: "one unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1, testAddr2},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1, testAddr2},
		},
		{
			name: "two unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr4, testAddr3},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1, testAddr2},
			},
			exp: []sdk.AccAddress{testAddr4, testAddr3, testAddr1, testAddr2},
		},
		{
			name: "three unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr3, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr4},
			},
			exp: []sdk.AccAddress{testAddr2, testAddr3, testAddr1, testAddr0, testAddr4},
		},
		{
			name: "two unaccepted three accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr4},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr3, testAddr1},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr4, testAddr2, testAddr3, testAddr1},
		},
		{
			name: "same address in both twice",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1, testAddr1},
			},
			exp: []sdk.AccAddress{testAddr1, testAddr1, testAddr1, testAddr1},
		},
	}

	// These shouldn't affect tests at all, but it's better to have
	// them set just in case, for some reason, they do.
	// But I didn't want to worry about them when defining the tests,
	// so I'm doing it here instead.
	for i, tc := range tests {
		if i%2 == 0 {
			tc.qr.Coins = coinMakerOK()
			tc.qr.Declined = true
		} else {
			tc.qr.Coins = coinMakerMulti()
			tc.qr.Declined = false
		}
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := makeCopyOfQuarantineRecord(tc.qr)
			actual := tc.qr.GetAllFromAddrs()
			assert.Equal(t, tc.exp, actual, "GetAllFromAddrs result")
			assert.Equal(t, orig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecord_AsQuarantinedFunds(t *testing.T) {
	testAddr0 := makeTestAddr("qrasqf", 0)
	testAddr1 := makeTestAddr("qrasqf", 1)
	testAddr2 := makeTestAddr("qrasqf", 2)

	tests := []struct {
		name     string
		qr       *QuarantineRecord
		toAddr   sdk.AccAddress
		expected *QuarantinedFunds
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "bad coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
		},
		{
			name: "empty coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
		},
		{
			name: "no to address",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: nil,
			expected: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "nil from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "empty from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1, testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String(), testAddr2.String()},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qrOrig := makeCopyOfQuarantineRecord(tc.qr)
			actual := tc.qr.AsQuarantinedFunds(tc.toAddr)
			assert.Equal(t, tc.expected, actual, "resulting QuarantinedFunds")
			assert.Equal(t, qrOrig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

// TODO[1046]: Test QuarantineRecordSuffixIndex.AddSuffixes
// TODO[1046]: Test QuarantineRecordSuffixIndex.RemoveSuffixes
// TODO[1046]: Test QuarantineRecordSuffixIndex.Simplify
