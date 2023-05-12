## [v0.46.10-pio-4](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.10-pio-4) - 2023-05-12

### Features

* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) Allow injection of send restrictions into the `x/bank` module.
* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) Create context helpers for bypassing the `x/quarantine` and `x/sanction` send restriction functions.

### Improvements

* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) Update the `x/quarantine` and `x/sanction` modules to use `SendRestrictionFn`s.

### API Breaking

* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) The `InputOutputCoins` function now takes in a single input (instead of slice of inputs).
* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) Remove the bank `SendKeeper` methods `SetQuarantineKeeper` and `SetSanctionKeeper`.
  Those have been refactored to `SendRestrictionFn`s.
* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) The `MintingRestrictionFn` type has been moved to the bank `types` package (from `keeper`).
* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) The `SendCoinsBypassQuarantine` function has been removed.
  It is replaced with useing `quarantine.WithBypass(...)` on the context being provided to `SendCoins`.
* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) Remove the bank keeper's `SetSendRestrictionsFunc`.
  Uses of it should be refactored into a `SendRestrictionFn` and provided to either `AppendSendRestriction` or `PrependSendRestriction`.

### State Machine Breaking

* [PR 568](https://github.com/provenance-io/cosmos-sdk/pull/568) `MsgMultiSend` now returns an error if there is more than one input.

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.10-pio-3...v0.46.10-pio-4
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.10..v0.46.10-pio-4
