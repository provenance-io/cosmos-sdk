package quarantine

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// TODO[1046]: Implement RegisterLegacyAminoCodec
	// Here's what gov has:
	// 	cdc.RegisterInterface((*Content)(nil), nil)
	//	legacy.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "cosmos-sdk/MsgSubmitProposal")
	//	legacy.RegisterAminoMsg(cdc, &MsgDeposit{}, "cosmos-sdk/MsgDeposit")
	//	legacy.RegisterAminoMsg(cdc, &MsgVote{}, "cosmos-sdk/MsgVote")
	//	legacy.RegisterAminoMsg(cdc, &MsgVoteWeighted{}, "cosmos-sdk/MsgVoteWeighted")
	//	cdc.RegisterConcrete(&TextProposal{}, "cosmos-sdk/TextProposal", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// TODO[1046]: Implement RegisterInterfaces
	// Here's what gov has:
	// 	registry.RegisterImplementations((*sdk.Msg)(nil),
	//		&MsgSubmitProposal{},
	//		&MsgVote{},
	//		&MsgVoteWeighted{},
	//		&MsgDeposit{},
	//	)
	//	registry.RegisterInterface(
	//		"cosmos.gov.v1beta1.Content",
	//		(*Content)(nil),
	//		&TextProposal{},
	//	)
	//
	//	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
}
