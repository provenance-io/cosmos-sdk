package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

func (s *TestSuite) TestOptIn() {
	addr0 := MakeTestAddr("optin", 0).String()

	tests := []struct {
		name   string
		msg    *quarantine.MsgOptIn
		expErr []string
	}{
		{
			name:   "bad address",
			msg:    &quarantine.MsgOptIn{ToAddress: "badbad"},
			expErr: []string{"decoding bech32 failed"},
		},
		{
			name: "okay",
			msg:  &quarantine.MsgOptIn{ToAddress: addr0},
		},
		{
			name: "repeat okay",
			msg:  &quarantine.MsgOptIn{ToAddress: addr0},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actResp, actErr := s.keeper.OptIn(s.stdlibCtx, tc.msg)
			AssertErrorContents(s.T(), actErr, tc.expErr, "OptIn error")
			if len(tc.expErr) == 0 {
				s.Assert().NotNil(actResp, "MsgOptInResponse")
				addr, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)
				if err == nil {
					isQ := s.keeper.IsQuarantinedAddr(s.sdkCtx, addr)
					s.Assert().True(isQ, "IsQuarantinedAddr")
				}
			}
		})
	}
}

func (s *TestSuite) TestOptOut() {
	addr0Acc := MakeTestAddr("optout", 0)
	addr0 := addr0Acc.String()
	addr1 := MakeTestAddr("oook", 1).String()

	// Setup, opt addr0 in so it can be opted out later.
	var err error
	testFunc := func() {
		err = s.keeper.SetOptIn(s.sdkCtx, addr0Acc)
	}
	s.Require().NotPanics(testFunc, "SetOptIn")
	s.Require().NoError(err, "SetOptIn")

	tests := []struct {
		name   string
		msg    *quarantine.MsgOptOut
		expErr []string
	}{
		{
			name:   "bad address",
			msg:    &quarantine.MsgOptOut{ToAddress: "badbad"},
			expErr: []string{"decoding bech32 failed"},
		},
		{
			name: "not opted in",
			msg:  &quarantine.MsgOptOut{ToAddress: addr1},
		},
		{
			name: "opted in",
			msg:  &quarantine.MsgOptOut{ToAddress: addr0},
		},
		{
			name: "opted in again",
			msg:  &quarantine.MsgOptOut{ToAddress: addr0},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actResp, actErr := s.keeper.OptOut(s.stdlibCtx, tc.msg)
			AssertErrorContents(s.T(), actErr, tc.expErr, "OptIn error")
			if len(tc.expErr) == 0 {
				s.Assert().NotNil(actResp, "MsgOptInResponse")
				addr, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)
				if err == nil {
					isQ := s.keeper.IsQuarantinedAddr(s.sdkCtx, addr)
					s.Assert().False(isQ, "IsQuarantinedAddr")
				}
			}
		})
	}
}

func (s *TestSuite) TestAccept() {
	makeAddr := func(index uint8) (sdk.AccAddress, string) {
		addr := MakeTestAddr("accept", index)
		return addr, addr.String()
	}
	makeFREvent := func(addr string, amt sdk.Coins) sdk.Event {
		rv, err := sdk.TypedEventToEvent(&quarantine.EventFundsReleased{
			ToAddress: addr,
			Coins:     amt,
		})
		s.Require().NoError(err, "TypedEventToEvent")
		return rv
	}
	addr0Acc, addr0 := makeAddr(0)
	addr1Acc, addr1 := makeAddr(1)
	addr2Acc, addr2 := makeAddr(2)

	// Set up some quarantined funds to 0 from 1 and to 0 from 2.
	amt1 := czt(s.T(), "5491book")
	amt2 := czt(s.T(), "8383tape")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, amt1, addr0Acc, addr1Acc), "AddQuarantinedCoins 0 1")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, amt2, addr0Acc, addr2Acc), "AddQuarantinedCoins 0 2")

	tests := []struct {
		name      string
		msg       *quarantine.MsgAccept
		expErr    []string
		expEvents sdk.Events
		expSend   *SentCoins
		expPerm   bool
	}{
		{
			name:   "bad to address",
			msg:    &quarantine.MsgAccept{ToAddress: "stillbad"},
			expErr: []string{"decoding bech32 failed", "invalid to address"},
		},
		{
			name: "bad first from",
			msg: &quarantine.MsgAccept{
				ToAddress:     addr0,
				FromAddresses: []string{"notgood"},
			},
			expErr: []string{"decoding bech32 failed", "invalid from address[0]"},
		},
		{
			name: "bad second from",
			msg: &quarantine.MsgAccept{
				ToAddress:     addr0,
				FromAddresses: []string{addr1, "notgood", addr2},
			},
			expErr: []string{"decoding bech32 failed", "invalid from address[1]"},
		},
		{
			name: "bad third from",
			msg: &quarantine.MsgAccept{
				ToAddress:     addr0,
				FromAddresses: []string{addr1, addr2, "notgood"},
			},
			expErr: []string{"decoding bech32 failed", "invalid from address[2]"},
		},
		{
			name: "nothing to accept",
			msg: &quarantine.MsgAccept{
				ToAddress:     addr2,
				FromAddresses: []string{addr0},
			},
			expEvents: sdk.Events{},
		},
		{
			name: "nothing to accept but permanent",
			msg: &quarantine.MsgAccept{
				ToAddress:     addr2,
				FromAddresses: []string{addr0},
				Permanent:     true,
			},
			expEvents: sdk.Events{},
			expPerm:   true,
		},
		{
			name: "funds to accept",
			msg: &quarantine.MsgAccept{
				ToAddress:     addr0,
				FromAddresses: []string{addr1},
			},
			expEvents: sdk.Events{makeFREvent(addr0, amt1)},
			expSend: &SentCoins{
				FromAddr: s.keeper.GetFundsHolder(),
				ToAddr:   addr0Acc,
				Amt:      MakeCopyOfCoins(amt1),
			},
		},
		{
			name: "funds to accept and permanent",
			msg: &quarantine.MsgAccept{
				ToAddress:     addr0,
				FromAddresses: []string{addr2},
				Permanent:     true,
			},
			expEvents: sdk.Events{makeFREvent(addr0, amt2)},
			expSend: &SentCoins{
				FromAddr: s.keeper.GetFundsHolder(),
				ToAddr:   addr0Acc,
				Amt:      MakeCopyOfCoins(amt2),
			},
			expPerm: true,
		},
		{
			name: "nothing to accept now not perm this time",
			msg: &quarantine.MsgAccept{
				ToAddress:     addr0,
				FromAddresses: []string{addr2},
				Permanent:     false,
			},
			expPerm: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			bKeeper := NewMockBankKeeper()
			qKeeper := s.keeper.WithBankKeeper(bKeeper)
			em := sdk.NewEventManager()
			ctx := sdk.WrapSDKContext(s.sdkCtx.WithEventManager(em))
			actResp, actErr := qKeeper.Accept(ctx, tc.msg)
			AssertErrorContents(s.T(), actErr, tc.expErr, "Accept error")
			if len(tc.expErr) == 0 {
				s.Assert().NotNil(actResp, "MsgAcceptResponse")
			}

			if tc.expEvents != nil {
				actEvents := em.Events()
				s.Assert().Equal(tc.expEvents, actEvents, "emitted events")
			}

			var expSends []*SentCoins
			if tc.expSend != nil {
				expSends = append(expSends, tc.expSend)
			}
			actSends := bKeeper.SentCoins
			s.Assert().Equal(expSends, actSends, "sends made")

			if tc.expPerm {
				toAddrAcc, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)
				if s.Assert().NoError(err, "toAddr to acc") {
					for _, fromAddr := range tc.msg.FromAddresses {
						fromAddrAcc, err := sdk.AccAddressFromBech32(fromAddr)
						if s.Assert().NoError(err, "fromAddr to acc") {
							actPerm := qKeeper.IsAutoAccept(s.sdkCtx, toAddrAcc, fromAddrAcc)
							s.Assert().True(actPerm, "IsAutoAccept")
						}
					}
				}
			}
		})
	}
}

func (s *TestSuite) TestDecline() {
	makeAddr := func(index uint8) (sdk.AccAddress, string) {
		addr := MakeTestAddr("decline", index)
		return addr, addr.String()
	}
	addr0Acc, addr0 := makeAddr(0)
	addr1Acc, addr1 := makeAddr(1)
	addr2Acc, addr2 := makeAddr(2)

	// Set up some quarantined funds to 0 from 1 and to 0 from 2.
	amt1 := czt(s.T(), "66route")
	amt2 := czt(s.T(), "55hagar")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, amt1, addr0Acc, addr1Acc), "AddQuarantinedCoins 0 1")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, amt2, addr0Acc, addr2Acc), "AddQuarantinedCoins 0 2")

	tests := []struct {
		name    string
		msg     *quarantine.MsgDecline
		expErr  []string
		expRec  *quarantine.QuarantineRecord
		expPerm bool
	}{
		{
			name:   "bad to address",
			msg:    &quarantine.MsgDecline{ToAddress: "stillbad"},
			expErr: []string{"decoding bech32 failed", "invalid to address"},
		},
		{
			name: "bad first from",
			msg: &quarantine.MsgDecline{
				ToAddress:     addr0,
				FromAddresses: []string{"notgood"},
			},
			expErr: []string{"decoding bech32 failed", "invalid from address[0]"},
		},
		{
			name: "bad second from",
			msg: &quarantine.MsgDecline{
				ToAddress:     addr0,
				FromAddresses: []string{addr1, "notgood", addr2},
			},
			expErr: []string{"decoding bech32 failed", "invalid from address[1]"},
		},
		{
			name: "bad third from",
			msg: &quarantine.MsgDecline{
				ToAddress:     addr0,
				FromAddresses: []string{addr1, addr2, "notgood"},
			},
			expErr: []string{"decoding bech32 failed", "invalid from address[2]"},
		},
		{
			name: "nothing to decline",
			msg: &quarantine.MsgDecline{
				ToAddress:     addr2,
				FromAddresses: []string{addr0},
			},
		},
		{
			name: "nothing to decline but permanent",
			msg: &quarantine.MsgDecline{
				ToAddress:     addr2,
				FromAddresses: []string{addr0},
				Permanent:     true,
			},
			expPerm: true,
		},
		{
			name: "funds to decline",
			msg: &quarantine.MsgDecline{
				ToAddress:     addr0,
				FromAddresses: []string{addr1},
			},
			expRec: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr1Acc),
				Coins:                   MakeCopyOfCoins(amt1),
				Declined:                true,
			},
		},
		{
			name: "funds to decline and permanent",
			msg: &quarantine.MsgDecline{
				ToAddress:     addr0,
				FromAddresses: []string{addr2},
				Permanent:     true,
			},
			expRec: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr2Acc),
				Coins:                   MakeCopyOfCoins(amt2),
				Declined:                true,
			},
			expPerm: true,
		},
		{
			name: "declined funds declined again but with perm",
			msg: &quarantine.MsgDecline{
				ToAddress:     addr0,
				FromAddresses: []string{addr1},
				Permanent:     true,
			},
			expRec: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr1Acc),
				Coins:                   MakeCopyOfCoins(amt1),
				Declined:                true,
			},
			expPerm: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := sdk.WrapSDKContext(s.sdkCtx.WithEventManager(em))
			actResp, actErr := s.keeper.Decline(ctx, tc.msg)
			AssertErrorContents(s.T(), actErr, tc.expErr, "Decline error")
			if len(tc.expErr) == 0 {
				s.Assert().NotNil(actResp, "MsgDeclineResponse")
			}

			if tc.expRec != nil {
				toAddrAcc, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)
				if s.Assert().NoError(err, "AccAddressFromBech32 toAddr") {
					actRec := s.keeper.GetQuarantineRecord(s.sdkCtx, toAddrAcc, tc.expRec.GetAllFromAddrs()...)
					s.Assert().Equal(tc.expRec, actRec, "resulting record")
				}
			}

			if tc.expPerm {
				toAddrAcc, err := sdk.AccAddressFromBech32(tc.msg.ToAddress)
				if s.Assert().NoError(err, "toAddr to acc") {
					for _, fromAddr := range tc.msg.FromAddresses {
						fromAddrAcc, err := sdk.AccAddressFromBech32(fromAddr)
						if s.Assert().NoError(err, "fromAddr to acc") {
							actPerm := s.keeper.IsAutoDecline(s.sdkCtx, toAddrAcc, fromAddrAcc)
							s.Assert().True(actPerm, "IsAutoDecline")
						}
					}
				}
			}
		})
	}
}

// TODO[1046]: TestUpdateAutoResponses()
