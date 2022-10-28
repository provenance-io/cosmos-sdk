package quarantine

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// TODO[1046]: Implement RegisterLegacyAminoCodec
	// Something like this:
	// 	legacy.RegisterAminoMsg(cdc, &MsgQuarantineOptIn{}, "cosmos-sdk/MsgQuarantineOptIn")
	//	legacy.RegisterAminoMsg(cdc, &MsgQuarantineOptOut{}, "cosmos-sdk/MsgQuarantineOptOut")
	//	legacy.RegisterAminoMsg(cdc, &MsgQuarantineAccept{}, "cosmos-sdk/MsgQuarantineAccept")
	//	legacy.RegisterAminoMsg(cdc, &MsgQuarantineDecline{}, "cosmos-sdk/MsgQuarantineDecline")
	//	legacy.RegisterAminoMsg(cdc, &MsgUpdateQuarantineAutoResponses{}, "cosmos-sdk/MsgUpdateQuarantineAutoResponses")
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgQuarantineOptIn{},
		&MsgQuarantineOptOut{},
		&MsgQuarantineAccept{},
		&MsgQuarantineDecline{},
		&MsgUpdateQuarantineAutoResponses{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/quarantine module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/quarantine and
	// defined at the application level.
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
