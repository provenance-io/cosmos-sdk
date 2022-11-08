package quarantine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewMsgOptIn(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("nmoi test addr 0"),
		testAddr("nmoi test addr 1"),
	}
	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected *MsgOptIn
	}{
		{
			name:     "addr 0",
			toAddr:   testAddrs[0],
			expected: &MsgOptIn{ToAddress: testAddrs[0].String()},
		},
		{
			name:     "addr 1",
			toAddr:   testAddrs[1],
			expected: &MsgOptIn{ToAddress: testAddrs[1].String()},
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: &MsgOptIn{ToAddress: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgOptIn(tc.toAddr)
			assert.Equal(t, tc.expected, actual, "NewMsgOptIn")
		})
	}
}

func TestMsgOptInValidateBasic(t *testing.T) {
	addr := testAddr("moivb test addr").String()
	tests := []struct {
		name          string
		addr          string
		expectedInErr []string
	}{
		{
			name:          "addr",
			addr:          addr,
			expectedInErr: nil,
		},
		{
			name:          "bad",
			addr:          "not an actual address",
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty",
			addr:          "",
			expectedInErr: []string{"invalid to address"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgOptIn{ToAddress: tc.addr}
			msg := MsgOptIn{ToAddress: tc.addr}
			err := msg.ValidateBasic()
			assertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgOptIn before and after")
		})
	}
}

func TestMsgOptInGetSigners(t *testing.T) {
	addr := testAddr("moigs test addr")
	tests := []struct {
		name     string
		addr     string
		expected []sdk.AccAddress
	}{
		{
			name:     "addr",
			addr:     addr.String(),
			expected: []sdk.AccAddress{addr},
		},
		{
			name:     "bad",
			addr:     "not an actual address",
			expected: []sdk.AccAddress{nil},
		},
		{
			name:     "empty",
			addr:     "",
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgOptIn{ToAddress: tc.addr}
			msg := MsgOptIn{ToAddress: tc.addr}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgOptIn before and after")
		})
	}
}

func TestNewMsgOptOut(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("nmoo test addr 0"),
		testAddr("nmoo test addr 1"),
	}
	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected *MsgOptOut
	}{
		{
			name:     "addr 0",
			toAddr:   testAddrs[0],
			expected: &MsgOptOut{ToAddress: testAddrs[0].String()},
		},
		{
			name:     "addr 1",
			toAddr:   testAddrs[1],
			expected: &MsgOptOut{ToAddress: testAddrs[1].String()},
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: &MsgOptOut{ToAddress: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgOptOut(tc.toAddr)
			assert.Equal(t, tc.expected, actual, "NewMsgOptOut")
		})
	}
}

func TestMsgOptOutValidateBasic(t *testing.T) {
	addr := testAddr("moovb test addr").String()
	tests := []struct {
		name          string
		addr          string
		expectedInErr []string
	}{
		{
			name:          "addr",
			addr:          addr,
			expectedInErr: nil,
		},
		{
			name:          "bad",
			addr:          "not an actual address",
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty",
			addr:          "",
			expectedInErr: []string{"invalid to address"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgOptOut{ToAddress: tc.addr}
			msg := MsgOptOut{ToAddress: tc.addr}
			err := msg.ValidateBasic()
			assertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgOptOut before and after")
		})
	}
}

func TestMsgOptOutGetSigners(t *testing.T) {
	addr := testAddr("moogs test addr")
	tests := []struct {
		name     string
		addr     string
		expected []sdk.AccAddress
	}{
		{
			name:     "addr",
			addr:     addr.String(),
			expected: []sdk.AccAddress{addr},
		},
		{
			name:     "bad",
			addr:     "not an actual address",
			expected: []sdk.AccAddress{nil},
		},
		{
			name:     "empty",
			addr:     "",
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgOptOut{ToAddress: tc.addr}
			msg := MsgOptOut{ToAddress: tc.addr}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgOptOut before and after")
		})
	}
}

func TestNewMsgAccept(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("nma test addr 0"),
		testAddr("nma test addr 1"),
	}
	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []string
		permanent bool
		expected  *MsgAccept
	}{
		{
			name:      "control",
			toAddr:    testAddrs[0],
			fromAddrs: []string{testAddrs[1].String()},
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     testAddrs[0].String(),
				FromAddresses: []string{testAddrs[1].String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []string{testAddrs[1].String()},
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     "",
				FromAddresses: []string{testAddrs[1].String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil fromAddrsStrs",
			toAddr:    testAddrs[1],
			fromAddrs: nil,
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     testAddrs[1].String(),
				FromAddresses: nil,
				Permanent:     false,
			},
		},
		{
			name:      "empty fromAddrsStrs",
			toAddr:    testAddrs[1],
			fromAddrs: []string{},
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     testAddrs[1].String(),
				FromAddresses: []string{},
				Permanent:     false,
			},
		},
		{
			name:      "three bad fromAddrsStrs",
			toAddr:    testAddrs[1],
			fromAddrs: []string{"one", "two", "three"},
			permanent: false,
			expected: &MsgAccept{
				ToAddress:     testAddrs[1].String(),
				FromAddresses: []string{"one", "two", "three"},
				Permanent:     false,
			},
		},
		{
			name:      "permanent",
			toAddr:    testAddrs[1],
			fromAddrs: []string{testAddrs[0].String()},
			permanent: true,
			expected: &MsgAccept{
				ToAddress:     testAddrs[1].String(),
				FromAddresses: []string{testAddrs[0].String()},
				Permanent:     true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgAccept(tc.toAddr, tc.fromAddrs, tc.permanent)
			assert.Equal(t, tc.expected, actual, "NewMsgAccept")
		})
	}
}

func TestMsgAcceptValidateBasic(t *testing.T) {
	testAddrs := []string{
		testAddr("mavb test addr 0").String(),
		testAddr("mavb test addr 1").String(),
		testAddr("mavb test addr 2").String(),
	}
	tests := []struct {
		name          string
		toAddr        string
		fromAddrs     []string
		permanent     bool
		expectedInErr []string
	}{
		{
			name:          "control",
			toAddr:        testAddrs[0],
			fromAddrs:     []string{testAddrs[1]},
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "permanent",
			toAddr:        testAddrs[0],
			fromAddrs:     []string{testAddrs[1]},
			permanent:     true,
			expectedInErr: nil,
		},
		{
			name:          "permanent no from addresses",
			toAddr:        testAddrs[2],
			fromAddrs:     []string{},
			permanent:     true,
			expectedInErr: []string{"at least one from address is required when permanent = true", "invalid value"},
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddrs:     []string{testAddrs[1]},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad to address",
			toAddr:        "this address isn't",
			fromAddrs:     []string{testAddrs[0]},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "nil from addresses",
			toAddr:        testAddrs[1],
			fromAddrs:     nil,
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "empty from addresses",
			toAddr:        testAddrs[1],
			fromAddrs:     []string{},
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "bad from address",
			toAddr:        testAddrs[0],
			fromAddrs:     []string{"this one is a tunic"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[0]"},
		},
		{
			name:          "bad third from address",
			toAddr:        testAddrs[0],
			fromAddrs:     []string{testAddrs[1], testAddrs[2], "Michael Jackson (he's bad)"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[2]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: makeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			err := msg.ValidateBasic()
			assertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgAccept before and after")
		})
	}
}

func TestMsgAcceptGetSigners(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("mags test addr 0"),
		testAddr("mags test addr 1"),
		testAddr("mags test addr 2"),
	}
	tests := []struct {
		name      string
		toAddr    string
		fromAddrs []string
		permanent bool
		expected  []sdk.AccAddress
	}{
		{
			name:      "control",
			toAddr:    testAddrs[0].String(),
			fromAddrs: []string{testAddrs[1].String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddrs[0]},
		},
		{
			name:      "permanent",
			toAddr:    testAddrs[0].String(),
			fromAddrs: []string{testAddrs[1].String()},
			permanent: true,
			expected:  []sdk.AccAddress{testAddrs[0]},
		},
		{
			name:      "empty to address",
			toAddr:    "",
			fromAddrs: []string{testAddrs[1].String()},
			permanent: false,
			expected:  []sdk.AccAddress{{}},
		},
		{
			name:      "bad to address",
			toAddr:    "this address isn't",
			fromAddrs: []string{testAddrs[0].String()},
			permanent: false,
			expected:  []sdk.AccAddress{nil},
		},
		{
			name:      "empty from addresses",
			toAddr:    testAddrs[1].String(),
			fromAddrs: []string{},
			permanent: false,
			expected:  []sdk.AccAddress{testAddrs[1]},
		},
		{
			name:      "two from addresses",
			toAddr:    testAddrs[2].String(),
			fromAddrs: []string{testAddrs[0].String(), testAddrs[1].String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddrs[2]},
		},
		{
			name:      "bad from address",
			toAddr:    testAddrs[0].String(),
			fromAddrs: []string{"this one is a tunic"},
			permanent: false,
			expected:  []sdk.AccAddress{testAddrs[0]},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: makeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgAccept before and after")
		})
	}
}

func TestNewMsgDecline(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("nmd test addr 0"),
		testAddr("nmd test addr 1"),
		testAddr("nmd test addr 2"),
	}
	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []string
		permanent bool
		expected  *MsgDecline
	}{
		{
			name:      "control",
			toAddr:    testAddrs[0],
			fromAddrs: []string{testAddrs[1].String()},
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     testAddrs[0].String(),
				FromAddresses: []string{testAddrs[1].String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []string{testAddrs[1].String()},
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     "",
				FromAddresses: []string{testAddrs[1].String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil fromAddrsStrs",
			toAddr:    testAddrs[1],
			fromAddrs: nil,
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     testAddrs[1].String(),
				FromAddresses: nil,
				Permanent:     false,
			},
		},
		{
			name:      "empty fromAddrsStrs",
			toAddr:    testAddrs[1],
			fromAddrs: []string{},
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     testAddrs[1].String(),
				FromAddresses: []string{},
				Permanent:     false,
			},
		},
		{
			name:      "three bad fromAddrsStrs",
			toAddr:    testAddrs[1],
			fromAddrs: []string{"one", "two", "three"},
			permanent: false,
			expected: &MsgDecline{
				ToAddress:     testAddrs[1].String(),
				FromAddresses: []string{"one", "two", "three"},
				Permanent:     false,
			},
		},
		{
			name:      "permanent",
			toAddr:    testAddrs[1],
			fromAddrs: []string{testAddrs[0].String()},
			permanent: true,
			expected: &MsgDecline{
				ToAddress:     testAddrs[1].String(),
				FromAddresses: []string{testAddrs[0].String()},
				Permanent:     true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgDecline(tc.toAddr, tc.fromAddrs, tc.permanent)
			assert.Equal(t, tc.expected, actual, "NewMsgDecline")
		})
	}
}

func TestMsgDeclineValidateBasic(t *testing.T) {
	testAddrs := []string{
		testAddr("mdvb test addr 0").String(),
		testAddr("mdvb test addr 1").String(),
		testAddr("mdvb test addr 2").String(),
	}
	tests := []struct {
		name          string
		toAddr        string
		fromAddrs     []string
		permanent     bool
		expectedInErr []string
	}{
		{
			name:          "control",
			toAddr:        testAddrs[0],
			fromAddrs:     []string{testAddrs[1]},
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "permanent",
			toAddr:        testAddrs[0],
			fromAddrs:     []string{testAddrs[1]},
			permanent:     true,
			expectedInErr: nil,
		},
		{
			name:          "permanent no from addresses",
			toAddr:        testAddrs[2],
			fromAddrs:     []string{},
			permanent:     true,
			expectedInErr: []string{"at least one from address is required when permanent = true", "invalid value"},
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddrs:     []string{testAddrs[1]},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad to address",
			toAddr:        "this address isn't",
			fromAddrs:     []string{testAddrs[0]},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "nil from addresses",
			toAddr:        testAddrs[1],
			fromAddrs:     nil,
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "empty from addresses",
			toAddr:        testAddrs[1],
			fromAddrs:     []string{},
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "bad from address",
			toAddr:        testAddrs[0],
			fromAddrs:     []string{"this one is a tunic"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[0]"},
		},
		{
			name:          "bad third from address",
			toAddr:        testAddrs[0],
			fromAddrs:     []string{testAddrs[1], testAddrs[2], "Michael Jackson (he's bad)"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[2]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: makeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			err := msg.ValidateBasic()
			assertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgDecline before and after")
		})
	}
}

func TestMsgDeclineGetSigners(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("mdgs test addr 0"),
		testAddr("mdgs test addr 1"),
		testAddr("mdgs test addr 2"),
	}
	tests := []struct {
		name      string
		toAddr    string
		fromAddrs []string
		permanent bool
		expected  []sdk.AccAddress
	}{
		{
			name:      "control",
			toAddr:    testAddrs[0].String(),
			fromAddrs: []string{testAddrs[1].String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddrs[0]},
		},
		{
			name:      "permanent",
			toAddr:    testAddrs[0].String(),
			fromAddrs: []string{testAddrs[1].String()},
			permanent: true,
			expected:  []sdk.AccAddress{testAddrs[0]},
		},
		{
			name:      "empty to address",
			toAddr:    "",
			fromAddrs: []string{testAddrs[1].String()},
			permanent: false,
			expected:  []sdk.AccAddress{{}},
		},
		{
			name:      "bad to address",
			toAddr:    "this address isn't",
			fromAddrs: []string{testAddrs[0].String()},
			permanent: false,
			expected:  []sdk.AccAddress{nil},
		},
		{
			name:      "empty from addresses",
			toAddr:    testAddrs[1].String(),
			fromAddrs: []string{},
			permanent: false,
			expected:  []sdk.AccAddress{testAddrs[1]},
		},
		{
			name:      "two from addresses",
			toAddr:    testAddrs[2].String(),
			fromAddrs: []string{testAddrs[0].String(), testAddrs[1].String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddrs[2]},
		},
		{
			name:      "bad from address",
			toAddr:    testAddrs[0].String(),
			fromAddrs: []string{"this one is a tunic"},
			permanent: false,
			expected:  []sdk.AccAddress{testAddrs[0]},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: makeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgDecline before and after")
		})
	}
}

func TestNewMsgUpdateAutoResponses(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("nmuar test addr 0"),
		testAddr("nmuar test addr 1"),
		testAddr("nmuar test addr 2"),
		testAddr("nmuar test addr 3"),
		testAddr("nmuar test addr 4"),
		testAddr("nmuar test addr 5"),
	}
	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		updates  []*AutoResponseUpdate
		expected *MsgUpdateAutoResponses
	}{
		{
			name:    "empty updates",
			toAddr:  testAddrs[0],
			updates: []*AutoResponseUpdate{},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddrs[0].String(),
				Updates:   []*AutoResponseUpdate{},
			},
		},
		{
			name:    "one update no to addr",
			toAddr:  nil,
			updates: []*AutoResponseUpdate{{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_ACCEPT}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: "",
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_ACCEPT}},
			},
		},
		{
			name:    "one update accept",
			toAddr:  testAddrs[1],
			updates: []*AutoResponseUpdate{{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_ACCEPT}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddrs[1].String(),
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_ACCEPT}},
			},
		},
		{
			name:    "one update decline",
			toAddr:  testAddrs[2],
			updates: []*AutoResponseUpdate{{FromAddress: testAddrs[1].String(), Response: AUTO_RESPONSE_DECLINE}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddrs[2].String(),
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddrs[1].String(), Response: AUTO_RESPONSE_DECLINE}},
			},
		},
		{
			name:    "one update unspecified",
			toAddr:  testAddrs[0],
			updates: []*AutoResponseUpdate{{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_UNSPECIFIED}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddrs[0].String(),
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_UNSPECIFIED}},
			},
		},
		{
			name:    "one update unspecified",
			toAddr:  testAddrs[0],
			updates: []*AutoResponseUpdate{{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_UNSPECIFIED}},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddrs[0].String(),
				Updates:   []*AutoResponseUpdate{{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_UNSPECIFIED}},
			},
		},
		{
			name:   "five updates",
			toAddr: testAddrs[0],
			updates: []*AutoResponseUpdate{
				{FromAddress: testAddrs[1].String(), Response: AUTO_RESPONSE_ACCEPT},
				{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_DECLINE},
				{FromAddress: testAddrs[3].String(), Response: AUTO_RESPONSE_ACCEPT},
				{FromAddress: testAddrs[4].String(), Response: AUTO_RESPONSE_UNSPECIFIED},
				{FromAddress: testAddrs[5].String(), Response: AUTO_RESPONSE_ACCEPT},
			},
			expected: &MsgUpdateAutoResponses{
				ToAddress: testAddrs[0].String(),
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1].String(), Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_DECLINE},
					{FromAddress: testAddrs[3].String(), Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[4].String(), Response: AUTO_RESPONSE_UNSPECIFIED},
					{FromAddress: testAddrs[5].String(), Response: AUTO_RESPONSE_ACCEPT},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewMsgUpdateAutoResponses(tc.toAddr, tc.updates)
			assert.Equal(t, tc.expected, actual, "NewMsgUpdateAutoResponses")
		})
	}
}

func TestMsgUpdateAutoResponsesValidateBasic(t *testing.T) {
	testAddrs := []string{
		testAddr("muarvb test addr 0").String(),
		testAddr("muarvb test addr 1").String(),
		testAddr("muarvb test addr 2").String(),
		testAddr("muarvb test addr 3").String(),
		testAddr("muarvb test addr 4").String(),
		testAddr("muarvb test addr 5").String(),
	}
	tests := []struct {
		name          string
		orig          MsgUpdateAutoResponses
		expectedInErr []string
	}{
		{
			name: "control accept",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1], Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "control decline",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[2], Response: AUTO_RESPONSE_DECLINE},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "control unspecified",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[3], Response: AUTO_RESPONSE_UNSPECIFIED},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "bad to address",
			orig: MsgUpdateAutoResponses{
				ToAddress: "not really that bad",
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1], Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "empty to address",
			orig: MsgUpdateAutoResponses{
				ToAddress: "",
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1], Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "nil updates",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates:   nil,
			},
			expectedInErr: []string{"invalid value", "no updates"},
		},
		{
			name: "empty updates",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates:   []*AutoResponseUpdate{},
			},
			expectedInErr: []string{"invalid value", "no updates"},
		},
		{
			name: "one update bad from address",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: "Okay, I'm bad again.", Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "invalid from address"},
		},
		{
			name: "one update empty from address",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: "", Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "invalid from address"},
		},
		{
			name: "one update negative resp",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1], Response: -1},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: -1"},
		},
		{
			name: "one update resp too large",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[2], Response: 900},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: 900"},
		},
		{
			name: "five updates third bad from address",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[2], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: "still not good", Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[4], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[5], Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 3", "invalid from address"},
		},
		{
			name: "five updates fourth empty from address",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[2], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[3], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: "", Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[5], Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 4", "invalid from address"},
		},
		{
			name: "five updates first negative resp",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1], Response: -88},
					{FromAddress: testAddrs[2], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[3], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[4], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[5], Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: -88"},
		},
		{
			name: "five update last resp too large",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0],
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[2], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[3], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[4], Response: AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddrs[5], Response: 55},
				},
			},
			expectedInErr: []string{"invalid update 5", "unknown auto-response value: 55"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgUpdateAutoResponses{
				ToAddress: tc.orig.ToAddress,
				Updates:   nil,
			}
			if tc.orig.Updates != nil {
				msg.Updates = []*AutoResponseUpdate{}
				for _, update := range tc.orig.Updates {
					msg.Updates = append(msg.Updates, &AutoResponseUpdate{
						FromAddress: update.FromAddress,
						Response:    update.Response,
					})
				}
			}
			err := msg.ValidateBasic()
			assertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, tc.orig, msg, "MsgUpdateAutoResponses before and after")
		})
	}
}

func TestMsgUpdateAutoResponsesGetSigners(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("muargs test addr 0"),
		testAddr("muargs test addr 1"),
		testAddr("muargs test addr 2"),
	}
	tests := []struct {
		name     string
		orig     MsgUpdateAutoResponses
		expected []sdk.AccAddress
	}{
		{
			name: "control",
			orig: MsgUpdateAutoResponses{
				ToAddress: testAddrs[0].String(),
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1].String(), Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{testAddrs[0]},
		},
		{
			name: "bad addr",
			orig: MsgUpdateAutoResponses{
				ToAddress: "bad bad bad",
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[2].String(), Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{nil},
		},
		{
			name: "empty addr",
			orig: MsgUpdateAutoResponses{
				ToAddress: "",
				Updates: []*AutoResponseUpdate{
					{FromAddress: testAddrs[1].String(), Response: AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgUpdateAutoResponses{
				ToAddress: tc.orig.ToAddress,
				Updates:   nil,
			}
			if tc.orig.Updates != nil {
				msg.Updates = []*AutoResponseUpdate{}
				for _, update := range tc.orig.Updates {
					msg.Updates = append(msg.Updates, &AutoResponseUpdate{
						FromAddress: update.FromAddress,
						Response:    update.Response,
					})
				}
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, tc.orig, msg, "MsgUpdateAutoResponses before and after")
		})
	}
}
