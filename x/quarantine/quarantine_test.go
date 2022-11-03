package quarantine

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
)

func testAddr(name string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(name)))
}

type coinMaker func() sdk.Coins

var coinMakerMap = map[string]coinMaker{
	"ok": func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("okcoin", 100)) },
	"multi": func() sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin("multicoina", 33), sdk.NewInt64Coin("multicoinb", 67))
	},
	"empty": func() sdk.Coins { return sdk.Coins{} },
	"nil":   func() sdk.Coins { return nil },
	"bad":   func() sdk.Coins { return sdk.Coins{sdk.Coin{Denom: "badcoin", Amount: sdk.NewInt(-1)}} },
}

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
	rv := &QuarantinedFunds{
		ToAddress:               orig.ToAddress,
		UnacceptedFromAddresses: nil,
		Coins:                   makeCopyOfCoins(orig.Coins),
		Declined:                orig.Declined,
	}
	if orig.UnacceptedFromAddresses != nil {
		rv.UnacceptedFromAddresses = make([]string, len(orig.UnacceptedFromAddresses))
		for i, addr := range orig.UnacceptedFromAddresses {
			rv.UnacceptedFromAddresses[i] = addr
		}
	}
	return rv
}

// makeCopyOfQuarantineRecord creates a copy of the provided QuarantineRecord and returns it.
func makeCopyOfQuarantineRecord(orig *QuarantineRecord) *QuarantineRecord {
	rv := &QuarantineRecord{
		UnacceptedFromAddresses: nil,
		Coins:                   makeCopyOfCoins(orig.Coins),
		Declined:                orig.Declined,
	}
	if orig.UnacceptedFromAddresses != nil {
		rv.UnacceptedFromAddresses = make([]sdk.AccAddress, len(orig.UnacceptedFromAddresses))
		for i, addr := range orig.UnacceptedFromAddresses {
			rv.UnacceptedFromAddresses[i] = make(sdk.AccAddress, len(addr))
			copy(rv.UnacceptedFromAddresses[i], addr)
		}
	}
	return rv
}

func TestNewQuarantinedFunds(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("nqf test addr 0"),
		testAddr("nqf test addr 1"),
	}
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
			toAddr:    testAddrs[0],
			fromAddrs: []sdk.AccAddress{testAddrs[1]},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 88)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 88)),
				Declined:                false,
			},
		},
		{
			name:      "declined true",
			toAddr:    testAddrs[0],
			fromAddrs: []sdk.AccAddress{testAddrs[1]},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 87)),
			declined:  true,
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 87)),
				Declined:                true,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []sdk.AccAddress{testAddrs[1]},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 86)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 86)),
				Declined:                false,
			},
		},
		{
			name:      "nil fromAddrs",
			toAddr:    testAddrs[0],
			fromAddrs: nil,
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
				Declined:                false,
			},
		},
		{
			name:      "empty fromAddrs",
			toAddr:    testAddrs[0],
			fromAddrs: []sdk.AccAddress{},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
				Declined:                false,
			},
		},
		{
			name:      "empty coins",
			toAddr:    testAddrs[0],
			fromAddrs: []sdk.AccAddress{testAddrs[1]},
			Coins:     sdk.Coins{},
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   sdk.Coins{},
				Declined:                false,
			},
		},
		{
			name:      "nil coins",
			toAddr:    testAddrs[0],
			fromAddrs: []sdk.AccAddress{testAddrs[1]},
			Coins:     nil,
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   nil,
				Declined:                false,
			},
		},
		{
			name:      "invalid coins",
			toAddr:    testAddrs[0],
			fromAddrs: []sdk.AccAddress{testAddrs[1]},
			Coins:     sdk.Coins{sdk.Coin{Denom: "bad", Amount: sdk.NewInt(-1)}},
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
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

func TestQuarantinedFundsAsQuarantineRecord(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("qfaqr test addr 0"),
		testAddr("qfaqr test addr 1"),
		testAddr("qfaqr test addr 2"),
	}
	tests := []struct {
		name     string
		qf       *QuarantinedFunds
		expected *QuarantineRecord
	}{
		{
			name: "control",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: []string{testAddrs[0].String()},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "no toAddr",
			qf: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddrs[0].String()},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "nil fromAddrs",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: nil,
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "empty fromAddrs",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "three fromAddrs",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: []string{testAddrs[0].String(), testAddrs[1].String(), testAddrs[2].String()},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0], testAddrs[1], testAddrs[2]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "declined true",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: []string{testAddrs[0].String()},
				Coins:                   coinMakerMap["ok"](),
				Declined:                true,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                true,
			},
		},
		{
			name: "multi coins",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: []string{testAddrs[0].String()},
				Coins:                   coinMakerMap["multi"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["multi"](),
				Declined:                false,
			},
		},
		{
			name: "empty coins",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: []string{testAddrs[0].String()},
				Coins:                   coinMakerMap["empty"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["empty"](),
				Declined:                false,
			},
		},
		{
			name: "nil coins",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: []string{testAddrs[0].String()},
				Coins:                   coinMakerMap["nil"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[1]},
				Coins:                   coinMakerMap["nil"](),
				Declined:                false,
			},
		},
		{
			name: "bad coins",
			qf: &QuarantinedFunds{
				ToAddress:               "toAddr",
				UnacceptedFromAddresses: []string{testAddrs[0].String()},
				Coins:                   coinMakerMap["bad"](),
				Declined:                false,
			},
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[1]},
				Coins:                   coinMakerMap["bad"](),
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qfOrig := makeCopyOfQuarantinedFunds(tc.qf)
			actual := tc.qf.AsQuarantineRecord()
			assert.Equal(t, tc.expected, actual, "resulting QuarantineRecord")
			assert.Equal(t, qfOrig, tc.qf, "QuarantinedFunds before and after")
		})
	}
}

func TestQuarantinedFundsValidate(t *testing.T) {
	testAddrs := []string{
		testAddr("qfv test addr 0").String(),
		testAddr("qfv test addr 1").String(),
		testAddr("qfv test addr 2").String(),
	}

	tests := []struct {
		name          string
		qf            *QuarantinedFunds
		expectedInErr []string
	}{
		{
			name: "control",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "declined true",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                true,
			},
			expectedInErr: nil,
		},
		{
			name: "bad to address",
			qf: &QuarantinedFunds{
				ToAddress:               "notgonnawork",
				UnacceptedFromAddresses: []string{testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "empty to address",
			qf: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "bad from address",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{"alsonotgood"},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "empty from address",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{""},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "nil from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: nil,
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required", "invalid value"},
		},
		{
			name: "empty from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required", "invalid value"},
		},
		{
			name: "two from addresses both good",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[1], testAddrs[2]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "two same from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[2], testAddrs[2]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"duplicate unaccepted from address", testAddrs[2]},
		},
		{
			name: "three from addresses same first last",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[1], testAddrs[2], testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"duplicate unaccepted from address", testAddrs[1]},
		},
		{
			name: "two from addresses first bad",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{"this is not an address", testAddrs[2]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "two from addresses last bad",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[1], "this is also bad"},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[1]"},
		},
		{
			name: "empty coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[1]},
				Coins:                   coinMakerMap["empty"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "nil coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[1]},
				Coins:                   coinMakerMap["nil"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "bad coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddrs[0],
				UnacceptedFromAddresses: []string{testAddrs[1]},
				Coins:                   coinMakerMap["bad"](),
				Declined:                false,
			},
			expectedInErr: []string{coinMakerMap["bad"]().String(), "amount is not positive"},
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
	testAddrs := []sdk.AccAddress{
		testAddr("nare test addr 0"),
		testAddr("nare test addr 1"),
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		resp     AutoResponse
		expected *AutoResponseEntry
	}{
		{
			name:     "accept",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			resp:     AUTO_RESPONSE_ACCEPT,
			expected: &AutoResponseEntry{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Response:    AUTO_RESPONSE_ACCEPT,
			},
		},
		{
			name:     "decline",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			resp:     AUTO_RESPONSE_DECLINE,
			expected: &AutoResponseEntry{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Response:    AUTO_RESPONSE_DECLINE,
			},
		},
		{
			name:     "unspecified",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			resp:     AUTO_RESPONSE_UNSPECIFIED,
			expected: &AutoResponseEntry{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Response:    AUTO_RESPONSE_UNSPECIFIED,
			},
		},
		{
			name:     "nil to address",
			toAddr:   nil,
			fromAddr: testAddrs[1],
			resp:     AUTO_RESPONSE_ACCEPT,
			expected: &AutoResponseEntry{
				ToAddress:   "",
				FromAddress: testAddrs[1].String(),
				Response:    AUTO_RESPONSE_ACCEPT,
			},
		},
		{
			name:     "nil from address",
			toAddr:   testAddrs[0],
			fromAddr: nil,
			resp:     AUTO_RESPONSE_DECLINE,
			expected: &AutoResponseEntry{
				ToAddress:   testAddrs[0].String(),
				FromAddress: "",
				Response:    AUTO_RESPONSE_DECLINE,
			},
		},
		{
			name:     "weird response",
			toAddr:   testAddrs[1],
			fromAddr: testAddrs[0],
			resp:     -3,
			expected: &AutoResponseEntry{
				ToAddress:   testAddrs[1].String(),
				FromAddress: testAddrs[0].String(),
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

func TestAutoResponseEntryValidate(t *testing.T) {
	testAddrs := []string{
		testAddr("arev test addr 0").String(),
		testAddr("arev test addr 1").String(),
	}

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
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: nil,
		},
		{
			name:          "decline",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: nil,
		},
		{
			name:          "unspecified",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: nil,
		},
		{
			name:          "bad to address",
			toAddr:        "notgonnawork",
			fromAddr:      testAddrs[1],
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddr:      testAddrs[1],
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad from address",
			toAddr:        testAddrs[0],
			fromAddr:      "alsonotgood",
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "empty from address",
			toAddr:        testAddrs[0],
			fromAddr:      "",
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "negative response",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			resp:          -1,
			expectedInErr: []string{"unknown auto-response value", "-1"},
		},
		{
			name:          "response too large",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
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

func TestAutoResponseUpdateValidate(t *testing.T) {
	testAddrs := []string{
		testAddr("arev test addr 0").String(),
		testAddr("arev test addr 1").String(),
	}

	tests := []struct {
		name          string
		fromAddr      string
		resp          AutoResponse
		qf            QuarantinedFunds
		expectedInErr []string
	}{
		{
			name:          "accept",
			fromAddr:      testAddrs[0],
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: nil,
		},
		{
			name:          "decline",
			fromAddr:      testAddrs[1],
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: nil,
		},
		{
			name:          "unspecified",
			fromAddr:      testAddrs[0],
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
			fromAddr:      testAddrs[1],
			resp:          -1,
			expectedInErr: []string{"unknown auto-response value", "-1"},
		},
		{
			name:          "response too large",
			fromAddr:      testAddrs[0],
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

func TestAutoResponseIsValid(t *testing.T) {
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

func TestAutoResponseIsAccept(t *testing.T) {
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

func TestAutoResponseIsDecline(t *testing.T) {
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
	testAddrs := []sdk.AccAddress{
		testAddr("nqr test addr 0"),
		testAddr("nqr test addr 1"),
		testAddr("nqr test addr 2"),
	}
	tests := []struct {
		name      string
		fromAddrs []string
		coins     sdk.Coins
		declined  bool
		expected  *QuarantineRecord
		expPanic  string
	}{
		{
			name:      "control",
			fromAddrs: []string{testAddrs[0].String()},
			coins:     coinMakerMap["ok"](),
			declined:  false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name:      "declined",
			fromAddrs: []string{testAddrs[0].String()},
			coins:     coinMakerMap["ok"](),
			declined:  true,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                true,
			},
		},
		{
			name:      "multi coins",
			fromAddrs: []string{testAddrs[0].String()},
			coins:     coinMakerMap["multi"](),
			declined:  false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["multi"](),
				Declined:                false,
			},
		},
		{
			name:      "empty coins",
			fromAddrs: []string{testAddrs[0].String()},
			coins:     coinMakerMap["empty"](),
			declined:  false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["empty"](),
				Declined:                false,
			},
		},
		{
			name:      "nil coins",
			fromAddrs: []string{testAddrs[0].String()},
			coins:     coinMakerMap["nil"](),
			declined:  false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["nil"](),
				Declined:                false,
			},
		},
		{
			name:      "bad coins",
			fromAddrs: []string{testAddrs[0].String()},
			coins:     coinMakerMap["bad"](),
			declined:  false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["bad"](),
				Declined:                false,
			},
		},
		{
			name:      "bad addr panics",
			fromAddrs: []string{"I'm a bad address"},
			coins:     coinMakerMap["ok"](),
			declined:  false,
			expPanic:  "TODO",
		},
		{
			name:      "empty addr string panics",
			fromAddrs: []string{""},
			coins:     coinMakerMap["ok"](),
			declined:  false,
			expPanic:  "TODO",
		},
		{
			name:      "two from addresses",
			fromAddrs: []string{testAddrs[0].String(), testAddrs[1].String()},
			coins:     coinMakerMap["ok"](),
			declined:  false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0], testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name:      "empty from addresses",
			fromAddrs: []string{},
			coins:     coinMakerMap["ok"](),
			declined:  false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name:      "nil from addresses",
			fromAddrs: nil,
			coins:     coinMakerMap["ok"](),
			declined:  false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *QuarantineRecord
			testFunc := func() {
				actual = NewQuarantineRecord(tc.fromAddrs, tc.coins, tc.declined)
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

func TestQuarantineRecordValidate(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("qrv test address 0"),
		testAddr("qrv test address 1"),
		testAddr("qrv test address 2"),
	}
	tests := []struct {
		name          string
		qr            *QuarantineRecord
		expectedInErr []string
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                true,
			},
			expectedInErr: nil,
		},
		{
			name: "multi coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["multi"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "empty coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["empty"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "nil coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["nil"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "bad coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0]},
				Coins:                   coinMakerMap["bad"](),
				Declined:                false,
			},
			expectedInErr: []string{coinMakerMap["bad"]().String(), "amount is not positive"},
		},
		{
			name: "nil from addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required"},
		},
		{
			name: "empty from addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required"},
		},
		{
			name: "two from addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0], testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "three from addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[0], testAddrs[1], testAddrs[2]},
				Coins:                   coinMakerMap["ok"](),
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

func TestQuarantineRecordIsZero(t *testing.T) {
	goodAddr := testAddr("qriz good address")
	tests := []struct {
		name     string
		qr       *QuarantineRecord
		expected bool
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{goodAddr},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{goodAddr},
				Coins:                   coinMakerMap["ok"](),
				Declined:                true,
			},
			expected: false,
		},
		{
			name: "multi coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{goodAddr},
				Coins:                   coinMakerMap["multi"](),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "empty coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{goodAddr},
				Coins:                   coinMakerMap["empty"](),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "nil coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{goodAddr},
				Coins:                   coinMakerMap["nil"](),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "bad coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{goodAddr},
				Coins:                   coinMakerMap["bad"](),
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

func TestQuarantineRecordAdd(t *testing.T) {
	moreCoinMakers := map[string]coinMaker{
		"empty":          coinMakerMap["empty"],
		"nil":            coinMakerMap["nil"],
		"0acorn":         func() sdk.Coins { return sdk.Coins{sdk.NewInt64Coin("acorn", 0)} },
		"50acorn":        func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)) },
		"32almond":       func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("almond", 32)) },
		"8acorn,9almond": func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)) },
	}

	tests := []struct {
		qrCoinKey  string
		addCoinKey string
		expected   sdk.Coins
	}{
		// empty
		{
			qrCoinKey:  "empty",
			addCoinKey: "empty",
			expected:   nil,
		},
		{
			qrCoinKey:  "empty",
			addCoinKey: "nil",
			expected:   nil,
		},
		{
			qrCoinKey:  "empty",
			addCoinKey: "0acorn",
			expected:   nil,
		},
		{
			qrCoinKey:  "empty",
			addCoinKey: "50acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  "empty",
			addCoinKey: "32almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  "empty",
			addCoinKey: "8acorn,9almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// nil
		{
			qrCoinKey:  "nil",
			addCoinKey: "empty",
			expected:   nil,
		},
		{
			qrCoinKey:  "nil",
			addCoinKey: "nil",
			expected:   nil,
		},
		{
			qrCoinKey:  "nil",
			addCoinKey: "0acorn",
			expected:   nil,
		},
		{
			qrCoinKey:  "nil",
			addCoinKey: "50acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  "nil",
			addCoinKey: "32almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  "nil",
			addCoinKey: "8acorn,9almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// 0acorn
		{
			qrCoinKey:  "0acorn",
			addCoinKey: "empty",
			expected:   nil,
		},
		{
			qrCoinKey:  "0acorn",
			addCoinKey: "nil",
			expected:   nil,
		},
		{
			qrCoinKey:  "0acorn",
			addCoinKey: "0acorn",
			expected:   nil,
		},
		{
			qrCoinKey:  "0acorn",
			addCoinKey: "50acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  "0acorn",
			addCoinKey: "32almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  "0acorn",
			addCoinKey: "8acorn,9almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// 50acorn
		{
			qrCoinKey:  "50acorn",
			addCoinKey: "empty",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  "50acorn",
			addCoinKey: "nil",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  "50acorn",
			addCoinKey: "0acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  "50acorn",
			addCoinKey: "50acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 100)),
		},
		{
			qrCoinKey:  "50acorn",
			addCoinKey: "32almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50), sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  "50acorn",
			addCoinKey: "8acorn,9almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 58), sdk.NewInt64Coin("almond", 9)),
		},

		// 32almond
		{
			qrCoinKey:  "32almond",
			addCoinKey: "empty",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  "32almond",
			addCoinKey: "nil",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  "32almond",
			addCoinKey: "0acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  "32almond",
			addCoinKey: "50acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50), sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  "32almond",
			addCoinKey: "32almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 64)),
		},
		{
			qrCoinKey:  "32almond",
			addCoinKey: "8acorn,9almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 41)),
		},

		// 8acorn,9almond
		{
			qrCoinKey:  "8acorn,9almond",
			addCoinKey: "empty",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  "8acorn,9almond",
			addCoinKey: "nil",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  "8acorn,9almond",
			addCoinKey: "0acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  "8acorn,9almond",
			addCoinKey: "50acorn",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 58), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  "8acorn,9almond",
			addCoinKey: "32almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 41)),
		},
		{
			qrCoinKey:  "8acorn,9almond",
			addCoinKey: "8acorn,9almond",
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 16), sdk.NewInt64Coin("almond", 18)),
		},
	}

	for _, declined := range []bool{false, true} {
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%q+%q=%q %t", tc.qrCoinKey, tc.addCoinKey, tc.expected.String(), declined), func(t *testing.T) {
				expected := QuarantineRecord{
					Coins:    tc.expected,
					Declined: declined,
				}
				qr := QuarantineRecord{
					Coins:    moreCoinMakers[tc.qrCoinKey](),
					Declined: declined,
				}
				addCoinsOrig := moreCoinMakers[tc.addCoinKey]()
				addCoins := moreCoinMakers[tc.addCoinKey]()
				qr.Add(addCoins...)
				assert.Equal(t, expected, qr, "QuarantineRecord after Add")
				assert.Equal(t, addCoinsOrig, addCoins, "Coins before and after")
			})
		}
	}
}

// TODO[1046]: Test QuarantineRecord.AcceptFromAddrs

func TestQuarantineRecordAsQuarantinedFunds(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("qrasqf test addr 0"),
		testAddr("qrasqf test addr 1"),
		testAddr("qrasqf test addr 2"),
	}
	tests := []struct {
		name     string
		qr       *QuarantineRecord
		toAddr   sdk.AccAddress
		expected *QuarantinedFunds
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			toAddr: testAddrs[0],
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                true,
			},
			toAddr: testAddrs[0],
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   coinMakerMap["ok"](),
				Declined:                true,
			},
		},
		{
			name: "bad coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[1]},
				Coins:                   coinMakerMap["bad"](),
				Declined:                false,
			},
			toAddr: testAddrs[0],
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   coinMakerMap["bad"](),
				Declined:                false,
			},
		},
		{
			name: "empty coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[1]},
				Coins:                   coinMakerMap["empty"](),
				Declined:                false,
			},
			toAddr: testAddrs[0],
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   coinMakerMap["empty"](),
				Declined:                false,
			},
		},
		{
			name: "no to address",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[1]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			toAddr: nil,
			expected: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddrs[1].String()},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "nil from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			toAddr: testAddrs[0],
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "empty from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			toAddr: testAddrs[0],
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
		},
		{
			name: "two from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddrs[1], testAddrs[2]},
				Coins:                   coinMakerMap["ok"](),
				Declined:                false,
			},
			toAddr: testAddrs[0],
			expected: &QuarantinedFunds{
				ToAddress:               testAddrs[0].String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerMap["ok"](),
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
