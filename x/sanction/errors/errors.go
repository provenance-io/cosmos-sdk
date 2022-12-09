package errors

import "cosmossdk.io/errors"

// sanctionCodespace is the codespace for all errors defined in sanction package
const sanctionCodespace = "sanction"

var ErrInvalidImmediateParams = errors.Register(sanctionCodespace, 2, "invalid immediate params")
