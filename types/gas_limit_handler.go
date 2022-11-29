package types

// GasLimitHandler optional, set's custom gas limit for provenance.
type GasLimitHandler func(ctx Context, defaultGasLimit uint64) (actualGasLimit uint64, err error)
