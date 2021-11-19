package types

// AdditionalMsgFeeHandler optional, which if set will charge additional fee for a msgType(if configured via governance)
type AdditionalMsgFeeHandler func(ctx Context, simulate bool) (coins Coins, err error)
