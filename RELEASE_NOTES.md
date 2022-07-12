# Provenance Specific Releases

## [v0.45.4-pio-1](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.45.4-pio-1) - 2022-04-22

### API Breaking Changes

* (x/bank) [\#11859](https://github.com/cosmos/cosmos-sdk/pull/11859) Move the SendEnabled information out of the Params and into the state store directly.
  The information can now be accessed using the BankKeeper.
  Setting can be done using MsgSetSendEnabled as a governance proposal.
  A SendEnabled query has been added to both GRPC and CLI.

### State Machine Breaking

* (x/bank) [\#11859](https://github.com/cosmos/cosmos-sdk/pull/11859) Move the SendEnabled information out of the Params and into the state store directly.

### Deprecated

* (x/bank) [\#11859](https://github.com/cosmos/cosmos-sdk/pull/11859) The Params.SendEnabled field is deprecated and unusable.

### Full Commit History

https://github.com/provenance-io/cosmos-sdk/compare/v0.45.3-pio-2...v0.45.4-pio-1
