package types

type AdditionalMsgFeeHandler func(ctx Context, simulate bool) (coins Coins, err error)

