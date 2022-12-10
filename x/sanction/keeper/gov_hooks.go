package keeper

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

var _ govtypes.GovHooks = Keeper{}

var msgSanctionTypeURL = sdk.MsgTypeURL(&sanction.MsgSanction{})
var msgUnsanctionTypeURL = sdk.MsgTypeURL(&sanction.MsgUnsanction{})

func IsModuleMsgURL(url string) bool {
	return url == msgSanctionTypeURL || url == msgUnsanctionTypeURL
}

// AfterProposalSubmission is called after proposal is submitted.
// If there's enough deposit, temporary entries are created.
func (k Keeper) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	k.handleProposal(ctx, proposalID)
}

// AfterProposalDeposit is called after a deposit is made.
// If there's enough deposit, temporary entries are created.
func (k Keeper) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
	k.handleProposal(ctx, proposalID)
}

// AfterProposalVote is called after a vote on a proposal is cast. This one does nothing.
func (k Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {}

// AfterProposalFailedMinDeposit is called when proposal fails to reach min deposit. This one does nothing.
func (k Keeper) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {}

// AfterProposalVotingPeriodEnded is called when proposal's finishes it's voting period.
// Cleans up temporary entries.
func (k Keeper) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	k.handleProposal(ctx, proposalID)
}

// handleProposal does what needs to be done in here with the proposal in question.
func (k Keeper) handleProposal(ctx sdk.Context, proposalID uint64) {
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		panic(fmt.Errorf("governance proposal not found with id %d", proposalID))
	}

	// TODO[1046]: Refactor this to account for proposals being deleted with they fail in some cases.
	for _, msg := range proposal.Messages {
		if IsModuleMsgURL(msg.TypeUrl) {
			switch proposal.Status {
			case govv1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD, govv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD:
				// If the deposit is over the minimum, add temporary entries for the addrs.
				deposit := sdk.Coins(proposal.TotalDeposit)
				minDeposit := k.getMinDeposit(ctx, msg)
				_, hasNeg := deposit.SafeSub(minDeposit...)
				if !hasNeg {
					addrs := k.getMsgAddresses(msg)
					var err error
					switch msg.TypeUrl {
					case msgSanctionTypeURL:
						err = k.AddTemporarySanction(ctx, proposalID, addrs...)
					case msgUnsanctionTypeURL:
						err = k.AddTemporaryUnsanction(ctx, proposalID, addrs...)
					}
					if err != nil {
						panic(err)
					}
				}
			case govv1.StatusPassed:
				// Delete all temporary entries for the addrs.
				// Since it's a passed vote, that supersedes any other temporary entries for each address.
				// The permanent updates should have happened when the message was executed at the end voting.
				addrs := k.getMsgAddresses(msg)
				k.DeleteTempEntries(ctx, addrs...)
			case govv1.StatusRejected, govv1.StatusFailed:
				// Delete only the temporary entries that were associated with this proposal.
				addrs := k.getMsgAddresses(msg)
				k.DeleteSpecificTempEntries(ctx, proposalID, addrs...)
			default:
				panic(fmt.Errorf("unknown governance proposal status: [%s]", proposal.Status))
			}
		}
	}
}

func (k Keeper) getMsgAddresses(msg *codectypes.Any) []sdk.AccAddress {
	switch msg.TypeUrl {
	case msgSanctionTypeURL:
		var msgSanction sanction.MsgSanction
		if err := k.cdc.UnpackAny(msg, &msgSanction); err != nil {
			panic(err)
		}
		addrs, err := toAccAddrs(msgSanction.Addresses)
		if err != nil {
			panic(err)
		}
		return addrs
	case msgUnsanctionTypeURL:
		var msgUnsanction sanction.MsgUnsanction
		if err := k.cdc.UnpackAny(msg, &msgUnsanction); err != nil {
			panic(err)
		}
		addrs, err := toAccAddrs(msgUnsanction.Addresses)
		if err != nil {
			panic(err)
		}
		return addrs
	}
	return nil
}

func (k Keeper) getMinDeposit(ctx sdk.Context, msg *codectypes.Any) sdk.Coins {
	switch msg.TypeUrl {
	case msgSanctionTypeURL:
		return k.GetImmediateSanctionMinDeposit(ctx)
	case msgUnsanctionTypeURL:
		return k.GetImmediateUnsanctionMinDeposit(ctx)
	}
	return sdk.Coins{}
}
