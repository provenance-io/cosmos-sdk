package errors

import "cosmossdk.io/errors"

// sanctionCodespace is the codespace for all errors defined in sanction package
const sanctionCodespace = "sanction"

var ErrInvalidParams = errors.Register(sanctionCodespace, 2, "invalid params")
var ErrUnsanctionableAddr = errors.Register(sanctionCodespace, 3, "address cannot be sanctioned")