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
// Otherwaise, asserts that there is an error, and that it contains each of the provided strings.
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

func TestNewQuarantinedFunds(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("nqf test addr 0"),
		testAddr("nqf test addr 1"),
	}
	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		Coins    sdk.Coins
		declined bool
		expected *QuarantinedFunds
	}{
		{
			name:     "control",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			Coins:    sdk.NewCoins(sdk.NewInt64Coin("rando", 88)),
			declined: false,
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       sdk.NewCoins(sdk.NewInt64Coin("rando", 88)),
				Declined:    false,
			},
		},
		{
			name:     "declined true",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			Coins:    sdk.NewCoins(sdk.NewInt64Coin("rando", 87)),
			declined: true,
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       sdk.NewCoins(sdk.NewInt64Coin("rando", 87)),
				Declined:    true,
			},
		},
		{
			name:     "nil toAddr",
			toAddr:   nil,
			fromAddr: testAddrs[1],
			Coins:    sdk.NewCoins(sdk.NewInt64Coin("rando", 86)),
			declined: false,
			expected: &QuarantinedFunds{
				ToAddress:   "",
				FromAddress: testAddrs[1].String(),
				Coins:       sdk.NewCoins(sdk.NewInt64Coin("rando", 86)),
				Declined:    false,
			},
		},
		{
			name:     "nil fromAddr",
			toAddr:   testAddrs[0],
			fromAddr: nil,
			Coins:    sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
			declined: false,
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: "",
				Coins:       sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
				Declined:    false,
			},
		},
		{
			name:     "empty coins",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			Coins:    sdk.Coins{},
			declined: false,
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       sdk.Coins{},
				Declined:    false,
			},
		},
		{
			name:     "nil coins",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			Coins:    nil,
			declined: false,
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       nil,
				Declined:    false,
			},
		},
		{
			name:     "invalid coins",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			Coins:    sdk.Coins{sdk.Coin{Denom: "bad", Amount: sdk.NewInt(-1)}},
			declined: false,
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       sdk.Coins{sdk.Coin{Denom: "bad", Amount: sdk.NewInt(-1)}},
				Declined:    false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewQuarantinedFunds(tc.toAddr, tc.fromAddr, tc.Coins, tc.declined)
			assert.Equal(t, tc.expected, actual, "NewQuarantinedFunds")
		})
	}
}

func TestQuarantinedFundsAsQuarantineRecord(t *testing.T) {
	tests := []struct {
		name     string
		toAddr   string
		fromAddr string
		coinKey  string
		declined bool
		expected *QuarantineRecord
	}{
		{
			name:     "control",
			toAddr:   "toAddr",
			fromAddr: "fromAddr",
			coinKey:  "ok",
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["ok"](),
				Declined: false,
			},
		},
		{
			name:     "no toAddr",
			toAddr:   "",
			fromAddr: "fromAddr",
			coinKey:  "ok",
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["ok"](),
				Declined: false,
			},
		},
		{
			name:     "no fromAddr",
			toAddr:   "toAddr",
			fromAddr: "",
			coinKey:  "ok",
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["ok"](),
				Declined: false,
			},
		},
		{
			name:     "declined true",
			toAddr:   "toAddr",
			fromAddr: "fromAddr",
			coinKey:  "ok",
			declined: true,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["ok"](),
				Declined: true,
			},
		},
		{
			name:     "multi coins",
			toAddr:   "toAddr",
			fromAddr: "fromAddr",
			coinKey:  "multi",
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["multi"](),
				Declined: false,
			},
		},
		{
			name:     "empty coins",
			toAddr:   "toAddr",
			fromAddr: "fromAddr",
			coinKey:  "empty",
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["empty"](),
				Declined: false,
			},
		},
		{
			name:     "nil coins",
			toAddr:   "toAddr",
			fromAddr: "fromAddr",
			coinKey:  "nil",
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["nil"](),
				Declined: false,
			},
		},
		{
			name:     "bad coins",
			toAddr:   "toAddr",
			fromAddr: "fromAddr",
			coinKey:  "bad",
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["bad"](),
				Declined: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qfOrig := QuarantinedFunds{
				ToAddress:   tc.toAddr,
				FromAddress: tc.fromAddr,
				Coins:       coinMakerMap[tc.coinKey](),
				Declined:    tc.declined,
			}
			qf := QuarantinedFunds{
				ToAddress:   tc.toAddr,
				FromAddress: tc.fromAddr,
				Coins:       coinMakerMap[tc.coinKey](),
				Declined:    tc.declined,
			}
			actual := qf.AsQuarantineRecord()
			assert.Equal(t, tc.expected, actual, "resulting QuarantineRecord")
			assert.Equal(t, qfOrig, qf, "QuarantinedFunds before and after")
		})
	}
}

func TestQuarantinedFundsValidate(t *testing.T) {
	testAddrs := []string{
		testAddr("qfv test addr 0").String(),
		testAddr("qfv test addr 1").String(),
	}

	tests := []struct {
		name          string
		toAddr        string
		fromAddr      string
		coinMkr       coinMaker
		declined      bool
		qf            QuarantinedFunds
		expectedInErr []string
	}{
		{
			name:          "control",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			coinMkr:       coinMakerMap["ok"],
			declined:      false,
			expectedInErr: nil,
		},
		{
			name:          "declined true",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			coinMkr:       coinMakerMap["ok"],
			declined:      true,
			expectedInErr: nil,
		},
		{
			name:          "bad to address",
			toAddr:        "notgonnawork",
			fromAddr:      testAddrs[1],
			coinMkr:       coinMakerMap["ok"],
			declined:      false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddr:      testAddrs[1],
			coinMkr:       coinMakerMap["ok"],
			declined:      false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad from address",
			toAddr:        testAddrs[0],
			fromAddr:      "alsonotgood",
			coinMkr:       coinMakerMap["ok"],
			declined:      false,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "empty from address",
			toAddr:        testAddrs[0],
			fromAddr:      "",
			coinMkr:       coinMakerMap["ok"],
			declined:      false,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "empty coins",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			coinMkr:       coinMakerMap["empty"],
			declined:      false,
			expectedInErr: nil,
		},
		{
			name:          "nil coins",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			coinMkr:       coinMakerMap["nil"],
			declined:      false,
			expectedInErr: nil,
		},
		{
			name:          "bad coins",
			toAddr:        testAddrs[0],
			fromAddr:      testAddrs[1],
			coinMkr:       coinMakerMap["bad"],
			declined:      false,
			expectedInErr: []string{coinMakerMap["bad"]().String(), "amount is not positive"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qfOrig := QuarantinedFunds{
				ToAddress:   tc.toAddr,
				FromAddress: tc.fromAddr,
				Coins:       tc.coinMkr(),
				Declined:    tc.declined,
			}
			qf := QuarantinedFunds{
				ToAddress:   tc.toAddr,
				FromAddress: tc.fromAddr,
				Coins:       tc.coinMkr(),
				Declined:    tc.declined,
			}
			err := qf.Validate()
			assertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, qfOrig, qf, "QuarantinedFunds before and after")
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
	tests := []struct {
		name     string
		coins    sdk.Coins
		declined bool
		expected *QuarantineRecord
	}{
		{
			name:     "control",
			coins:    coinMakerMap["ok"](),
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["ok"](),
				Declined: false,
			},
		},
		{
			name:     "declined",
			coins:    coinMakerMap["ok"](),
			declined: true,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["ok"](),
				Declined: true,
			},
		},
		{
			name:     "multi coins",
			coins:    coinMakerMap["multi"](),
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["multi"](),
				Declined: false,
			},
		},
		{
			name:     "empty coins",
			coins:    coinMakerMap["empty"](),
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["empty"](),
				Declined: false,
			},
		},
		{
			name:     "nil coins",
			coins:    coinMakerMap["nil"](),
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["nil"](),
				Declined: false,
			},
		},
		{
			name:     "bad coins",
			coins:    coinMakerMap["bad"](),
			declined: false,
			expected: &QuarantineRecord{
				Coins:    coinMakerMap["bad"](),
				Declined: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewQuarantineRecord(tc.coins, tc.declined)
			assert.Equal(t, tc.expected, actual, "NewQuarantineRecord")
		})
	}
}

func TestQuarantineRecordValidate(t *testing.T) {
	tests := []struct {
		name          string
		coinMkr       coinMaker
		declined      bool
		expectedInErr []string
	}{
		{
			name:          "control",
			coinMkr:       coinMakerMap["ok"],
			declined:      false,
			expectedInErr: nil,
		},
		{
			name:          "declined",
			coinMkr:       coinMakerMap["ok"],
			declined:      true,
			expectedInErr: nil,
		},
		{
			name:          "multi coins",
			coinMkr:       coinMakerMap["multi"],
			declined:      false,
			expectedInErr: nil,
		},
		{
			name:          "empty coins",
			coinMkr:       coinMakerMap["empty"],
			declined:      false,
			expectedInErr: nil,
		},
		{
			name:          "nil coins",
			coinMkr:       coinMakerMap["nil"],
			declined:      false,
			expectedInErr: nil,
		},
		{
			name:          "bad coins",
			coinMkr:       coinMakerMap["bad"],
			declined:      false,
			expectedInErr: []string{coinMakerMap["bad"]().String(), "amount is not positive"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qrOrig := QuarantineRecord{
				Coins:    tc.coinMkr(),
				Declined: tc.declined,
			}
			qr := QuarantineRecord{
				Coins:    tc.coinMkr(),
				Declined: tc.declined,
			}
			err := qr.Validate()
			assertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, qrOrig, qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecordIsZero(t *testing.T) {
	tests := []struct {
		name     string
		coinMkr  coinMaker
		declined bool
		expected bool
	}{
		{
			name:     "control",
			coinMkr:  coinMakerMap["ok"],
			declined: false,
			expected: false,
		},
		{
			name:     "declined",
			coinMkr:  coinMakerMap["ok"],
			declined: true,
			expected: false,
		},
		{
			name:     "multi coins",
			coinMkr:  coinMakerMap["multi"],
			declined: false,
			expected: false,
		},
		{
			name:     "empty coins",
			coinMkr:  coinMakerMap["empty"],
			declined: false,
			expected: true,
		},
		{
			name:     "nil coins",
			coinMkr:  coinMakerMap["nil"],
			declined: false,
			expected: true,
		},
		{
			name:     "bad coins",
			coinMkr:  coinMakerMap["bad"],
			declined: false,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qrOrig := QuarantineRecord{
				Coins:    tc.coinMkr(),
				Declined: tc.declined,
			}
			qr := QuarantineRecord{
				Coins:    tc.coinMkr(),
				Declined: tc.declined,
			}
			actual := qr.IsZero()
			assert.Equal(t, tc.expected, actual, "IsZero()")
			assert.Equal(t, qrOrig, qr, "QuarantineRecord before and after")
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

func TestQuarantineRecordAsQuarantinedFunds(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("qrasqf test addr 0"),
		testAddr("qrasqf test addr 1"),
	}
	tests := []struct {
		name     string
		coinMkr  coinMaker
		declined bool
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		expected *QuarantinedFunds
	}{
		{
			name:     "control",
			coinMkr:  coinMakerMap["ok"],
			declined: false,
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       coinMakerMap["ok"](),
				Declined:    false,
			},
		},
		{
			name:     "declined",
			coinMkr:  coinMakerMap["ok"],
			declined: true,
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       coinMakerMap["ok"](),
				Declined:    true,
			},
		},
		{
			name:     "bad coins",
			coinMkr:  coinMakerMap["bad"],
			declined: false,
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       coinMakerMap["bad"](),
				Declined:    false,
			},
		},
		{
			name:     "empty coins",
			coinMkr:  coinMakerMap["empty"],
			declined: false,
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: testAddrs[1].String(),
				Coins:       coinMakerMap["empty"](),
				Declined:    false,
			},
		},
		{
			name:     "no to address",
			coinMkr:  coinMakerMap["ok"],
			declined: false,
			toAddr:   nil,
			fromAddr: testAddrs[1],
			expected: &QuarantinedFunds{
				ToAddress:   "",
				FromAddress: testAddrs[1].String(),
				Coins:       coinMakerMap["ok"](),
				Declined:    false,
			},
		},
		{
			name:     "no from address",
			coinMkr:  coinMakerMap["ok"],
			declined: false,
			toAddr:   testAddrs[0],
			fromAddr: nil,
			expected: &QuarantinedFunds{
				ToAddress:   testAddrs[0].String(),
				FromAddress: "",
				Coins:       coinMakerMap["ok"](),
				Declined:    false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qrOrig := QuarantineRecord{
				Coins:    tc.coinMkr(),
				Declined: tc.declined,
			}
			qr := QuarantineRecord{
				Coins:    tc.coinMkr(),
				Declined: tc.declined,
			}
			actual := qr.AsQuarantinedFunds(tc.toAddr, tc.fromAddr)
			assert.Equal(t, tc.expected, actual, "resulting QuarantinedFunds")
			assert.Equal(t, qrOrig, qr, "QuarantineRecord before and after")
		})
	}
}
