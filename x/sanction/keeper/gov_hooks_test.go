package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type GovHooksTestSuite struct {
	BaseTestSuite
}

func (s *GovHooksTestSuite) SetupTest() {
	s.BaseSetup()
}

func TestGovHooksTestSuite(t *testing.T) {
	suite.Run(t, new(GovHooksTestSuite))
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalSubmission() {
	// Since this just calls proposalGovHook, all we should test in here is
	// that the proposalGovHook function was called for the given gov prop id.
	// So just mock up the gov keeper to return a proposal of interest, but with a bad status.
	// Hopefully the panic message that causes is unique to the proposalGovHook function.
	// We test that the call panics with the expected message.
	// We also test that GetProposal was called as expected.

	govPropID := uint64(3982)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 5555,
	}

	expPanic := "unknown governance proposal status: [5555]"
	testFunc := func() {
		s.Keeper.AfterProposalSubmission(s.SdkCtx, govPropID)
	}
	s.GovKeeper.GetProposalCalls = nil
	testutil.RequirePanicsWithMessage(s.T(), expPanic, testFunc, "AfterProposalSubmission")
	actualCalls := s.GovKeeper.GetProposalCalls
	if s.Assert().Len(actualCalls, 1, "number of calls made to GetProposal") {
		s.Assert().Equal(int(govPropID), int(actualCalls[0]), "the proposal requested to GetProposal")
	}
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalDeposit() {
	// Since this just calls proposalGovHook, all we should test in here is
	// that the proposalGovHook function was called for the given gov prop id.
	// So just mock up the gov keeper to return a proposal of interest, but with a bad status.
	// Hopefully the panic message that causes is unique to the proposalGovHook function.
	// We test that the call panics with the expected message.
	// We also test that GetProposal was called as expected.

	govPropID := uint64(5994)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 4434,
	}

	expPanic := "unknown governance proposal status: [4434]"
	testFunc := func() {
		s.Keeper.AfterProposalDeposit(s.SdkCtx, govPropID, sdk.AccAddress("this doesn't matter"))
	}
	s.GovKeeper.GetProposalCalls = nil
	testutil.RequirePanicsWithMessage(s.T(), expPanic, testFunc, "AfterProposalDeposit")
	actualCalls := s.GovKeeper.GetProposalCalls
	if s.Assert().Len(actualCalls, 1, "number of calls made to GetProposal") {
		s.Assert().Equal(int(govPropID), int(actualCalls[0]), "the proposal requested to GetProposal")
	}
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalVote() {
	// This one shouldn't do anything. So again, set it up to panic like the others,
	// but make sure it doesn't panic and that no calls were made to GetProposal

	govPropID := uint64(6370)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 2411,
	}

	testFunc := func() {
		s.Keeper.AfterProposalVote(s.SdkCtx, govPropID, sdk.AccAddress("this doesn't matter either"))
	}
	s.GovKeeper.GetProposalCalls = nil
	s.Require().NotPanics(testFunc, "AfterProposalVote")
	actualCalls := s.GovKeeper.GetProposalCalls
	s.Require().Nil(actualCalls, "calls made to GetProposal")
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalFailedMinDeposit() {
	// Since this just calls proposalGovHook, all we should test in here is
	// that the proposalGovHook function was called for the given gov prop id.
	// So just mock up the gov keeper to return a proposal of interest, but with a bad status.
	// Hopefully the panic message that causes is unique to the proposalGovHook function.
	// We test that the call panics with the expected message.
	// We also test that GetProposal was called as expected.

	govPropID := uint64(2111)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 3275,
	}

	expPanic := "unknown governance proposal status: [3275]"
	testFunc := func() {
		s.Keeper.AfterProposalFailedMinDeposit(s.SdkCtx, govPropID)
	}
	s.GovKeeper.GetProposalCalls = nil
	testutil.RequirePanicsWithMessage(s.T(), expPanic, testFunc, "AfterProposalFailedMinDeposit")
	actualCalls := s.GovKeeper.GetProposalCalls
	if s.Assert().Len(actualCalls, 1, "number of calls made to GetProposal") {
		s.Assert().Equal(int(govPropID), int(actualCalls[0]), "the proposal requested to GetProposal")
	}
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalVotingPeriodEnded() {
	// Since this just calls proposalGovHook, all we should test in here is
	// that the proposalGovHook function was called for the given gov prop id.
	// So just mock up the gov keeper to return a proposal of interest, but with a bad status.
	// Hopefully the panic message that causes is unique to the proposalGovHook function.
	// We test that the call panics with the expected message.
	// We also test that GetProposal was called as expected.

	govPropID := uint64(4041)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 99,
	}

	expPanic := "unknown governance proposal status: [99]"
	testFunc := func() {
		s.Keeper.AfterProposalVotingPeriodEnded(s.SdkCtx, govPropID)
	}
	s.GovKeeper.GetProposalCalls = nil
	testutil.RequirePanicsWithMessage(s.T(), expPanic, testFunc, "AfterProposalVotingPeriodEnded")
	actualCalls := s.GovKeeper.GetProposalCalls
	if s.Assert().Len(actualCalls, 1, "number of calls made to GetProposal") {
		s.Assert().Equal(int(govPropID), int(actualCalls[0]), "the proposal requested to GetProposal")
	}
}

// TODO[1046]: proposalGovHook(ctx sdk.Context, proposalID uint64)
func (s *GovHooksTestSuite) TestKeeper_proposalGovHook() {
	lastPropID := uint64(0)
	nextPropID := func() uint64 {
		lastPropID += 1
		return lastPropID
	}
	// When using lastPropID directly in test definition, things got out of sync.
	// Basically, nextPropID was being called for all the tests before lastPropID was being used.
	// Having it returned by a func like this fixed that though.
	curPropID := func() uint64 {
		return lastPropID
	}

	addr1 := sdk.AccAddress("1st_hooks_test_addr")
	addr2 := sdk.AccAddress("2nd_hooks_test_addr")
	addr3 := sdk.AccAddress("3rd_hooks_test_addr")
	addr4 := sdk.AccAddress("4th_hooks_test_addr")
	addr5 := sdk.AccAddress("5th_hooks_test_addr")
	addr6 := sdk.AccAddress("6th_hooks_test_addr")

	nonEmptyState := func(govPropID uint64) *sanction.GenesisState {
		return &sanction.GenesisState{
			Params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("nesanct", 17)),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("neusanct", 23)),
			},
			SanctionedAddresses: []string{addr3.String(), addr4.String()},
			TemporaryEntries: []*sanction.TemporaryEntry{
				newTempEntry(addr5, govPropID, true),
				newTempEntry(addr6, govPropID, true),
			},
		}
	}

	updateParamsAny := s.NewAny(&sanction.MsgUpdateParams{
		Params: &sanction.Params{
			ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("ismdcoin", 59)),
			ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("iumdcoin", 86)),
		},
		Authority: "yadda yadda",
	})
	sanctionAny := s.NewAny(&sanction.MsgSanction{
		Addresses: []string{addr1.String(), addr2.String()},
		Authority: "nospendy",
	})
	unsanctionAny := s.NewAny(&sanction.MsgUnsanction{
		Addresses: []string{addr1.String(), addr3.String()},
		Authority: "spendyagainy",
	})
	otherAny1 := s.NewAny(&govv1.MsgExecLegacyContent{
		Content:   sanctionAny,
		Authority: "legacywrapping",
	})
	otherAny2 := s.NewAny(&govv1.MsgExecLegacyContent{
		Content:   sanctionAny,
		Authority: "legacywrapping",
	})

	tests := []struct {
		name       string
		proposalID uint64
		iniState   *sanction.GenesisState
		proposal   *govv1.Proposal
		expState   *sanction.GenesisState
		expPanic   []string
	}{
		// prop status unknown -> panic if a message of interest is in the proposal
		{
			name:       "unknown prop status on other message",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny1},
				Status:   482,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "unknown prop status on update params",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{updateParamsAny},
				Status:   23948,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "unknown prop status on sanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{sanctionAny},
				Status:   45897439,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"unknown governance proposal status: [45897439]"},
		},
		{
			name:       "unknown prop status on unsanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{unsanctionAny},
				Status:   640958983,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"unknown governance proposal status: [640958983]"},
		},
		{
			name:       "unknown prop status on three ignorable messages",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny1, updateParamsAny, otherAny2},
				Status:   39834323,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "unknown prop status on three messages second is sanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny2, sanctionAny, otherAny1},
				Status:   49494994,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"unknown governance proposal status: [49494994]"},
		},
		{
			name:       "unknown prop status on three messages second is unsanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny2, unsanctionAny, otherAny1},
				Status:   2525252,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"unknown governance proposal status: [2525252]"},
		},

		// prop status unspecified -> panic if a message of interest is in the proposal
		{
			name:       "unspecified prop status on other message",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny2},
				Status:   govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "unspecified prop status on update params",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{updateParamsAny},
				Status:   govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "unspecified prop status on sanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{sanctionAny},
				Status:   govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"unknown governance proposal status: [PROPOSAL_STATUS_UNSPECIFIED]"},
		},
		{
			name:       "unspecified prop status on unsanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{unsanctionAny},
				Status:   govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"unknown governance proposal status: [PROPOSAL_STATUS_UNSPECIFIED]"},
		},
		{
			name:       "unknown prop status on three ignorable messages",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny1, updateParamsAny, otherAny2},
				Status:   govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "unknown prop status on three messages second is sanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny2, sanctionAny, otherAny1},
				Status:   govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"unknown governance proposal status: [PROPOSAL_STATUS_UNSPECIFIED]"},
		},
		{
			name:       "unknown prop status on three messages second is unsanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny1, unsanctionAny, otherAny2},
				Status:   govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"unknown governance proposal status: [PROPOSAL_STATUS_UNSPECIFIED]"},
		},

		// prop passed -> nothing happens in any case.
		{
			name:       "passed on other message",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny1},
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "passed on update params",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{updateParamsAny},
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "passed on sanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{sanctionAny},
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "passed on unsanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{unsanctionAny},
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "passed on three ignorable messages",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny2, updateParamsAny, otherAny1},
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "passed on three messages last is sanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny1, updateParamsAny, sanctionAny},
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "passed on three messages last is unsanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{otherAny2, updateParamsAny, unsanctionAny},
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		},
		{
			name:       "passed on three messages sanction other unsanction",
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: []*codectypes.Any{sanctionAny, otherAny2, unsanctionAny},
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		},
	}

	// Prop status situations to test:
	// Done: unknown prop status -> panic
	// Done: prop status = ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED -> panic
	// Done: prop passed -> nothing happens (make sure state isn't changed)
	// prop rejected ->  it's temp entries are deleted.
	// prop failed ->  it's temp entries are deleted.
	// prop not found -> it's temp entries are deleted.
	// prop StatusDepositPeriod -> more stuff
	// prop StatusVotingPeriod -> more stuff

	// more stuff:
	//   min deposit = 0 -> nothing happens
	//   deposit < min deposit -> nothing happens
	//   deposit == min deposit -> temp entries added.
	//   deposit > min deposit -> temp entries added.

	// There are 5 message types I'd like to involve in these tests:
	// 1) MsgSanction -> the above stuff, temp entries = sanction
	// 2) MsgUnsanction -> the above stuff, temp entries = unsanction
	// 3) MsgUpdateParams -> the above stuff, but nothing happens on any of it.
	// 4) MsgExecLegacyContent with a MsgSanction -> the above stuff, but nothing happens on any of it.
	// 5) MsgExecLegacyContent with a MsgUnsanction -> the above stuff, but nothing happens on any of it.

	// Then, just to make it complicated, I want to test combos of those messages:
	// A), [1, 2] = all the stuff from just 1 and just 2, in order though, so have a temp unsanction from 2 overwrite a temp sanction added by 1.
	// B), [2, 1] = same as A, but the temp sanction overwrites the temp unsanction.
	// C), [1, 1] = all sanctions from both
	// D), [2, 2] = all unsanctions from both
	// E), [3, 1] = 1 still happens
	// F), [1, 3] = 1 still happens
	// G), [1, 5] = make 5 undo 1 and make sure it is ignored.
	// G), [2, 4] = make 4 undo 2 and make sure it is ignored.
	// H), [1, 1, 1, 1, 1] = make sure they all happen.

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ClearState()
			if tc.iniState != nil {
				s.Require().NotPanics(func() {
					s.Keeper.InitGenesis(s.SdkCtx, tc.iniState)
				}, "InitGenesis")
			}

			s.GovKeeper.GetProposalCalls = nil
			if tc.proposal != nil {
				s.GovKeeper.GetProposalReturns[tc.proposal.Id] = *tc.proposal
			}

			testFunc := func() {
				s.Keeper.OnlyTestsProposalGovHook(s.SdkCtx, tc.proposalID)
			}
			testutil.RequirePanicContents(s.T(), tc.expPanic, testFunc, "proposalGovHook(%d)", tc.proposalID)

			getPropCalls := s.GovKeeper.GetProposalCalls
			if s.Assert().Len(getPropCalls, 1, "number of calls made to GetProposal") {
				// doing it this way because a failure message from an .Equal on two []uint64 slices shows
				// the values in hex. Since they're decimal in test definition, this is just easier.
				s.Assert().Equal(int(tc.proposalID), int(getPropCalls[0]), "gov prop id provided to GetProposal")
			}

			if tc.expState != nil {
				s.ExportAndCheck(tc.expState)
			}
		})
	}
}

func (s *GovHooksTestSuite) TestKeeper_isModuleGovHooksMsgURL() {
	tests := []struct {
		url string
		exp bool
	}{
		{exp: true, url: sdk.MsgTypeURL(&sanction.MsgSanction{})},
		{exp: true, url: sdk.MsgTypeURL(&sanction.MsgUnsanction{})},
		{exp: false, url: ""},
		{exp: false, url: "     "},
		{exp: false, url: "something random"},
		{exp: false, url: sdk.MsgTypeURL(&sanction.MsgUpdateParams{})},
		{exp: false, url: sdk.MsgTypeURL(&govv1.MsgExecLegacyContent{})},
		{exp: false, url: "cosmos.sanction.v1beta1.MsgSanction"},
		{exp: false, url: "/cosmos.sanction.v1beta1.MsgSanctio"},
		{exp: false, url: "/cosmos.sanction.v1beta1.MsgSanction "},
		{exp: false, url: " /cosmos.sanction.v1beta1.MsgSanction"},
		{exp: false, url: "/cosmos.sanction.v1beta1.MsgSanction2"},
	}

	for _, tc := range tests {
		name := tc.url
		if name == "" {
			name = "empty"
		}
		if strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("spaces x %d", len(name))
		}
		s.Run(name, func() {
			var actual bool
			testFunc := func() {
				actual = s.Keeper.OnlyTestsIsModuleGovHooksMsgURL(tc.url)
			}
			s.Require().NotPanics(testFunc, "isModuleGovHooksMsgURL(%q)", tc.url)
			s.Assert().Equal(tc.exp, actual, "isModuleGovHooksMsgURL(%q) result", tc.url)
		})
	}
}

func (s *GovHooksTestSuite) TestKeeper_getMsgAddresses() {
	addr1 := sdk.AccAddress("1_good_addr_for_test")
	addr2 := sdk.AccAddress("2_good_addr_for_test")
	addr3 := sdk.AccAddress("3_good_addr_for_test")
	addr4 := sdk.AccAddress("4_good_addr_for_test")
	addr5 := sdk.AccAddress("5_good_addr_for_test")
	addr6 := sdk.AccAddress("6_good_addr_for_test")

	tests := []struct {
		name     string
		msg      *codectypes.Any
		exp      []sdk.AccAddress
		expPanic []string
	}{
		// Tests for things outside the switch.
		{
			name: "nil",
			msg:  nil,
			exp:  nil,
		},
		{
			name: "type url is empty but content is MsgSanction",
			msg: s.CustomAny(nil, &sanction.MsgSanction{
				Addresses: []string{addr1.String()},
				Authority: "whatever",
			}),
			exp: nil,
		},
		{
			name: "type url is empty but content is MsgUnsanction",
			msg: s.CustomAny(nil, &sanction.MsgUnsanction{
				Addresses: []string{addr1.String()},
				Authority: "whatever",
			}),
			exp: nil,
		},
		{
			name: "MsgUpdateParams",
			msg: s.NewAny(&sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("pcoin", 1)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("pcoin", 2)),
				},
				Authority: "whatever",
			}),
			exp: nil,
		},
		{
			name: "MsgExecLegacyContent with a MsgSanction in it",
			msg: s.NewAny(&govv1.MsgExecLegacyContent{
				Content: s.NewAny(&sanction.MsgSanction{
					Addresses: []string{addr1.String()},
					Authority: "whatever2",
				}),
				Authority: "whatever",
			}),
			exp: nil,
		},
		{
			name: "MsgExecLegacyContent with a MsgUnsanction in it",
			msg: s.NewAny(&govv1.MsgExecLegacyContent{
				Content: s.NewAny(&sanction.MsgUnsanction{
					Addresses: []string{addr1.String()},
					Authority: "whatever2",
				}),
				Authority: "whatever",
			}),
			exp: nil,
		},

		// Tests for the MsgSanction case.
		{
			name: "MsgSanction nil addrs",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: nil,
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{},
		},
		{
			name: "MsgSanction empty addrs",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{},
		},
		{
			name: "MsgSanction one addr good",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{addr1.String()},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{addr1},
		},
		{
			name: "MsgSanction one addr bad",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this1isnotgood"},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "MsgSanction six addrs all good",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{addr1, addr2, addr3, addr4, addr5, addr6},
		},
		{
			name: "MsgSanction six addrs bad xxxx",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{
					"this1isalsobad",
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "MsgSanction six addrs bad third",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					"this1isthethirdbadone",
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[2]", "decoding bech32 failed"},
		},
		{
			name: "MsgSanction six addrs bad sixth",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					"another1thatisnotgood",
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[5]", "decoding bech32 failed"},
		},
		{
			name: "type is MsgSanction but content is not",
			msg: s.CustomAny(&sanction.MsgSanction{}, &govv1.MsgVote{
				ProposalId: 5,
				Voter:      addr1.String(),
				Option:     govv1.OptionNo,
				Metadata:   "I do not know what is going on",
			}),
			expPanic: []string{"no registered implementations of type *sanction.MsgSanction"},
		},

		// Tests for the MsgUnsanction case.
		{
			name: "MsgUnsanction nil addrs",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: nil,
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{},
		},
		{
			name: "MsgUnsanction empty addrs",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{},
		},
		{
			name: "MsgUnsanction one addr good",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{addr1.String()},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{addr1},
		},
		{
			name: "MsgUnsanction one addr bad",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{"this1isnotgood"},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "MsgUnsanction six addrs all good",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{addr1, addr2, addr3, addr4, addr5, addr6},
		},
		{
			name: "MsgUnsanction six addrs bad xxxx",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{
					"this1isalsobad",
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "MsgUnsanction six addrs bad third",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					"this1isthethirdbadone",
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[2]", "decoding bech32 failed"},
		},
		{
			name: "MsgUnsanction six addrs bad sixth",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					"another1thatisnotgood",
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[5]", "decoding bech32 failed"},
		},
		{
			name: "type is MsgUnsanction but content is not",
			msg: s.CustomAny(&sanction.MsgUnsanction{}, &govv1.MsgVote{
				ProposalId: 5,
				Voter:      addr1.String(),
				Option:     govv1.OptionNo,
				Metadata:   "I do not know what is going on",
			}),
			expPanic: []string{"no registered implementations of type *sanction.MsgUnsanction"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual []sdk.AccAddress
			testFunc := func() {
				actual = s.Keeper.OnlyTestsGetMsgAddresses(tc.msg)
			}
			testutil.AssertPanicContents(s.T(), tc.expPanic, testFunc, "getMsgAddresses")
			s.Assert().Equal(tc.exp, actual, "getMsgAddresses result")
		})
	}
}

func (s *GovHooksTestSuite) TestKeeper_getImmediateMinDeposit() {
	origSanctMin := sanction.DefaultImmediateSanctionMinDeposit
	origUnsanctMin := sanction.DefaultImmediateUnsanctionMinDeposit
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origSanctMin
		sanction.DefaultImmediateUnsanctionMinDeposit = origUnsanctMin
	}()
	sanction.DefaultImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("dsanct", 3))
	sanction.DefaultImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("dunsanct", 7))

	paramSanctMin := sdk.NewCoins(sdk.NewInt64Coin("psanct", 5))
	paramUnsanctMin := sdk.NewCoins(sdk.NewInt64Coin("punsanct", 10))

	tests := []struct {
		name string
		msg  *codectypes.Any
		exp  sdk.Coins // expected for either case.
		expd sdk.Coins // expected when getting from defaults.
		expp sdk.Coins // expected when getting from params.
	}{
		{
			name: "nil",
			msg:  nil,
			exp:  sdk.Coins{},
		},
		{
			name: "MsgSanction",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: nil,
				Authority: "whatever",
			}),
			expd: sanction.DefaultImmediateSanctionMinDeposit,
			expp: paramSanctMin,
		},
		{
			name: "MsgUnsanction",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: nil,
				Authority: "whatever",
			}),
			expd: sanction.DefaultImmediateUnsanctionMinDeposit,
			expp: paramUnsanctMin,
		},
		{
			name: "MsgExecLegacyContent with a MsgSanction",
			msg: s.NewAny(&govv1.MsgExecLegacyContent{
				Content: s.NewAny(&sanction.MsgSanction{
					Addresses: []string{"some dumb addr", "another unsavory addr"},
					Authority: "whatever2",
				}),
				Authority: "whatever",
			}),
			exp: sdk.Coins{},
		},
		{
			name: "MsgExecLegacyContent with a MsgUnsanction",
			msg: s.NewAny(&govv1.MsgExecLegacyContent{
				Content: s.NewAny(&sanction.MsgUnsanction{
					Addresses: []string{"some dumb addr", "another unsavory addr"},
					Authority: "whatever2",
				}),
				Authority: "whatever",
			}),
			exp: sdk.Coins{},
		},
		{
			name: "MsgUpdateParams",
			msg: s.NewAny(&sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("qcoin", 72)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("qcoin", 91)),
				},
				Authority: "whatever",
			}),
			exp: sdk.Coins{},
		},
	}

	// Delete the params so that the defaults are used.
	testutil.RequireNotPanicsNoError(s.T(), func() error {
		return s.Keeper.SetParams(s.SdkCtx, nil)
	}, "SetParams(nil)")

	for _, tc := range tests {
		s.Run(tc.name+" from defaults", func() {
			expected := tc.expd
			if expected == nil {
				expected = tc.exp
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.Keeper.OnlyTestsGetImmediateMinDeposit(s.SdkCtx, tc.msg)
			}
			s.Require().NotPanics(testFunc, "getImmediateMinDeposit")
			s.Assert().Equal(expected, actual, "getImmediateMinDeposit result")
		})
	}

	// Now, set the params appropriately.
	testutil.RequireNotPanicsNoError(s.T(), func() error {
		return s.Keeper.SetParams(s.SdkCtx, &sanction.Params{
			ImmediateSanctionMinDeposit:   paramSanctMin,
			ImmediateUnsanctionMinDeposit: paramUnsanctMin,
		})
	}, "SetParams with values")

	for _, tc := range tests {
		s.Run(tc.name+" from params", func() {
			expected := tc.expp
			if expected == nil {
				expected = tc.exp
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.Keeper.OnlyTestsGetImmediateMinDeposit(s.SdkCtx, tc.msg)
			}
			s.Require().NotPanics(testFunc, "getImmediateMinDeposit")
			s.Assert().Equal(expected, actual, "getImmediateMinDeposit result")
		})
	}
}
