package sanction

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ sdk.Msg = &MsgSanction{}

func (m MsgSanction) ValidateBasic() error {
	// TODO[1046]: Implement MsgSanction.ValidateBasic
	panic("not implemented")
}

func (m MsgSanction) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgUnsanction{}

func (m MsgUnsanction) ValidateBasic() error {
	// TODO[1046]: Implement MsgUnsanction.ValidateBasic
	panic("not implemented")
}

func (m MsgUnsanction) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgImmediateSanction{}

func (m MsgImmediateSanction) ValidateBasic() error {
	// TODO[1046]: Implement MsgImmediateSanction.ValidateBasic
	panic("not implemented")
}

func (m MsgImmediateSanction) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Proposer)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgImmediateUnsanction{}

func (m MsgImmediateUnsanction) ValidateBasic() error {
	// TODO[1046]: Implement MsgImmediateUnsanction.ValidateBasic
	panic("not implemented")
}

func (m MsgImmediateUnsanction) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Proposer)
	return []sdk.AccAddress{addr}
}
