## [v0.46.13-pio-2](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.13-pio-2) - 2023-08-11

### Features

* [#578](https://github.com/provenance-io/cosmos-sdk/pull/578) Add `binary_version` to the `NodeInfo` object returned by status command.
* [#577](https://github.com/provenance-io/cosmos-sdk/pull/577) Add an injectable `GetLockedCoinsFn` to the bank module.

### Improvements

* [#580](https://github.com/provenance-io/cosmos-sdk/pull/580) Add SpendableBalances query CLI command.
* [#581](https://github.com/provenance-io/cosmos-sdk/pull/581) For `MsgMultiSend` and `InputOutputCoins`, allow many inputs when there's a single output.

### Bug Fixes

* [#582](https://github.com/provenance-io/cosmos-sdk/pull/582) Prevent locked coins from being delegated. Coins locked in a vesting account can still be delegated though.
* [#582](https://github.com/provenance-io/cosmos-sdk/pull/582) Staking simulations can no longer try to delegate coins locked outside of vesting.

### API Breaking

* [#581](https://github.com/provenance-io/cosmos-sdk/pull/581) The `InputOutputCoins` once again takes in multiple inputs. It returns an error if there are multiple inputs and multiple outputs.
* [#582](https://github.com/provenance-io/cosmos-sdk/pull/582) The `SimulateFromSeed` function now also returns the last block time simulated.

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.13-pio-1...v0.46.13-pio-2
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.13..v0.46.13-pio-2

