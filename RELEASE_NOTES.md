## [v0.46.4-pio-1](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.4-pio-1) - 2022-11-15

### Improvements

* Bring in Cosmos-SDK [v0.46.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.4) changes [#362](https://github.com/provenance-io/cosmos-sdk/pull/362).
* (server) [#362](https://github.com/provenance-io/cosmos-sdk/pull/362) Change the default for the re-added start command --iavl-disable-fastnode flag back to true to match the config default.

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.3-pio-4...v0.46.4-pio-1
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.4...v0.46.4-pio-1

---

## [v0.46.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.5) - 2022-11-17

This release introduces a number of serious bug fixes and improvements. Notably, an upgrade to Tendermint [v0.34.23](https://github.com/tendermint/tendermint/releases/tag/v0.34.23).

If you are planning to migrate to v0.46, please use `v0.46.5`. All releases prior to `v0.46.5` are [retracted](https://go.dev/ref/mod#go-mod-file-retract) and **must NOT be used** (`go get` directly upgrades the SDK version to `>= v0.46.5` thanks to the retraction, current builds are not affected).

If your chain's state has coin metadata, an issue has been discovered in the bank module coin metadata migration. This issue is fixed in `v0.46.5`.

* If your chain is already on v0.46 using `<= v0.46.4` and has coin metadata, a **coordinated upgrade** to `v0.46.5` is required.
    * Use the helper function `Migrate_V0464_To_V0465` for migrating a chain **already on v0.46 with versions <=v0.46.4** to the latest v0.46.5 correct state.
* If your chain is already on v0.46 using `<= v0.46.4` but has no coin metadata, this release is **non-breaking**.

Moreover, serious issues have been found in the group module. These issues are fixed in `v0.46.5`.

* If you use the group module, upgrade to `v0.46.5` **immediately**. A **coordinated upgrade** to `v0.46.5` is required.

When a chain is already using `<= v0.46.4`, but has no coin metadata and no group module, this release is **non-breaking**.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.4...v0.46.5

**NOTE**: The changes mentioned in `v0.46.3` are **still** required:

```go
# Chains must add the following to their go.mod for the application:
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

### Features

* (x/bank) [#13891](https://github.com/cosmos/cosmos-sdk/pull/13891) Provide a helper function `Migrate_V0464_To_V0465` for migrating a chain **already on v0.46 with versions <=v0.46.4** to the latest v0.46.5 correct state.

### Improvements

* [#13826](https://github.com/cosmos/cosmos-sdk/pull/13826) Support custom `GasConfig` configuration for applications.
* (deps) Bump Tendermint version to [v0.34.23](https://github.com/tendermint/tendermint/releases/tag/v0.34.23).

### State Machine Breaking

* (x/group) [#13876](https://github.com/cosmos/cosmos-sdk/pull/13876) Fix group MinExecutionPeriod that is checked on execution now, instead of voting period end.

### API Breaking Changes

* (x/group) [#13876](https://github.com/cosmos/cosmos-sdk/pull/13876) Add `GetMinExecutionPeriod` method on DecisionPolicy interface.

### Bug Fixes

* (x/group) [#13869](https://github.com/cosmos/cosmos-sdk/pull/13869) Group members weight must be positive and a finite number.
* (x/bank) [#13821](https://github.com/cosmos/cosmos-sdk/pull/13821) Fix bank store migration of coin metadata.
* (x/group) [#13808](https://github.com/cosmos/cosmos-sdk/pull/13808) Fix propagation of message events to the current context in `EndBlocker`.
* (x/gov) [#13728](https://github.com/cosmos/cosmos-sdk/pull/13728) Fix propagation of message events to the current context in `EndBlocker`.
* (store) [#13803](https://github.com/cosmos/cosmos-sdk/pull/13803) Add an error log if IAVL set operation failed.
* [#13861](https://github.com/cosmos/cosmos-sdk/pull/13861) Allow `_` characters in tx event queries, i.e. `GetTxsEvent`.

