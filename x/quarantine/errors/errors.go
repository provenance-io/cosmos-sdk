package errors

import "cosmossdk.io/errors"

// quarantineCodespace is the codespace for all errors defined in quarantine package
const quarantineCodespace = "group"

var (
	ErrInvalidValue = errors.Register(quarantineCodespace, 2, "invalid value")
)
