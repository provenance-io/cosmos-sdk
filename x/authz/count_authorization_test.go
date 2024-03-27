package authz_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/authz"
)

func TestCountAuthorization(t *testing.T) {
	tests := []struct {
		name         string
		msgType      string
		count        int32
		expVBErr     string
		expAccept    bool
		expDelete    bool
		expUpdated   bool
		expAcceptErr string
	}{
		{
			name:         "negative one",
			msgType:      "/cosmos.bank.v1beta1.MsgSend",
			count:        -1,
			expVBErr:     "allowed authorizations must be greater than 0: invalid request",
			expAcceptErr: "allowed authorizations must be greater than 0: unauthorized",
		},
		{
			name:         "zero",
			msgType:      "weird",
			count:        0,
			expVBErr:     "allowed authorizations must be greater than 0: invalid request",
			expAcceptErr: "allowed authorizations must be greater than 0: unauthorized",
		},
		{name: "one", msgType: "something.else", count: 1, expAccept: true, expDelete: true},
		{name: "two", msgType: "another/url", count: 2, expAccept: true, expUpdated: true},
		{name: "five", msgType: "cinco", count: 5, expAccept: true, expUpdated: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var authorization *authz.CountAuthorization
			testNew := func() {
				authorization = authz.NewCountAuthorization(tc.msgType, tc.count)
			}
			require.NotPanics(t, testNew, "NewCountAuthorization(%q, %d)", tc.msgType, tc.count)
			require.NotNil(t, authorization, "NewCountAuthorization(%q, %d) result", tc.msgType, tc.count)

			var actMsgType string
			testMsgType := func() {
				actMsgType = authorization.MsgTypeURL()
			}
			if assert.NotPanics(t, testMsgType, "MsgTypeURL") {
				assert.Equal(t, tc.msgType, actMsgType, "MsgTypeURL value")
			}

			var actVBErr error
			testVB := func() {
				actVBErr = authorization.ValidateBasic()
			}
			if assert.NotPanics(t, testVB, "ValidateBasic") {
				if len(tc.expVBErr) > 0 {
					assert.EqualError(t, actVBErr, tc.expVBErr, "ValidateBasic error")
				} else {
					assert.NoError(t, actVBErr, "ValidateBasic error")
				}
			}

			var actAcceptResp authz.AcceptResponse
			var actAcceptErr error
			testAccept := func() {
				actAcceptResp, actAcceptErr = authorization.Accept(context.Background(), nil)
			}
			if assert.NotPanics(t, testAccept, "Accept") {
				if len(tc.expAcceptErr) > 0 {
					assert.EqualError(t, actAcceptErr, tc.expAcceptErr, "Accept error")
				} else {
					assert.NoError(t, actAcceptErr, "Accept error")
				}
				assert.Equal(t, tc.expAccept, actAcceptResp.Accept, "AcceptResponse.Accept")
				assert.Equal(t, tc.expDelete, actAcceptResp.Delete, "AcceptResponse.Delete")
				switch {
				case !tc.expUpdated:
					assert.Nil(t, actAcceptResp.Updated, "AcceptResponse.Updated")
				case assert.NotNil(t, actAcceptResp.Updated, "AcceptResponse.Updated"):
					actAuth, isCount := actAcceptResp.Updated.(*authz.CountAuthorization)
					if assert.True(t, isCount, "AcceptResponse.Updated should be a %T, but is a %T", authorization, actAcceptResp.Updated) {
						assert.Equal(t, tc.msgType, actAuth.Msg, "updated msg type")
						assert.Equal(t, int(tc.count-1), int(actAuth.AllowedAuthorizations), "updated msg count")
					}
				}
			}
		})
	}
}
