## [v0.46.13-pio-1](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.13-pio-1) - 2023-06-09

### Improvements

* [#572](https://github.com/provenance-io/cosmos-sdk/pull/572) Bring in Cosmos-SDK [v0.46.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.11), [v0.46.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.12), and [v0.46.13](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.13) changes.
* [#572](https://github.com/provenance-io/cosmos-sdk/pull/572) Add the better legacy gov prop error handling fix back in.

### Bug Fixes

* [#572](https://github.com/provenance-io/cosmos-sdk/pull/572) Fix group sims from creating multiple group policies with the same address.

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.10-pio-4...v0.46.13-pio-1
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.13..v0.46.13-pio-1

---

## [v0.46.13](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.13) - 2023-06-08

### Features

* (snapshots) [#16060](https://github.com/cosmos/cosmos-sdk/pull/16060) Support saving and restoring snapshot locally.
* (baseapp) [#16290](https://github.com/cosmos/cosmos-sdk/pull/16290) Add circuit breaker setter in baseapp.
* (x/group) [#16191](https://github.com/cosmos/cosmos-sdk/pull/16191) Add EventProposalPruned event to group module whenever a proposal is pruned.

### Improvements

* (deps) [#15973](https://github.com/cosmos/cosmos-sdk/pull/15973) Bump CometBFT to [v0.34.28](https://github.com/cometbft/cometbft/blob/v0.34.28/CHANGELOG.md#v03428).
* (store) [#15683](https://github.com/cosmos/cosmos-sdk/pull/15683) `rootmulti.Store.CacheMultiStoreWithVersion` now can handle loading archival states that don't persist any of the module stores the current state has.
* (simapp) [#15903](https://github.com/cosmos/cosmos-sdk/pull/15903) Add `AppStateFnWithExtendedCbs` with moduleStateCb callback function to allow access moduleState. Note, this function is present in `simtestutil` from `v0.47.2+`.
* (gov) [#15979](https://github.com/cosmos/cosmos-sdk/pull/15979) Improve gov error message when failing to convert v1 proposal to v1beta1.
* (server) [#16061](https://github.com/cosmos/cosmos-sdk/pull/16061) Add Comet bootstrap command.
* (store) [#16067](https://github.com/cosmos/cosmos-sdk/pull/16067) Add local snapshots management commands.
* (baseapp) [#16193](https://github.com/cosmos/cosmos-sdk/pull/16193) Add `Close` method to `BaseApp` for custom app to cleanup resource in graceful shutdown.

### Bug Fixes

* Fix [barberry](https://forum.cosmos.network/t/cosmos-sdk-security-advisory-barberry/10825) security vulnerability.
* (cli) [#16312](https://github.com/cosmos/cosmos-sdk/pull/16312) Allow any addresses in `client.ValidatePromptAddress`.
* (store/iavl) [#15717](https://github.com/cosmos/cosmos-sdk/pull/15717) Upstream error on empty version (this change was present on all version but v0.46).

---

## [v0.46.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.12) - 2023-04-04

### Features

* (x/groups) [#14879](https://github.com/cosmos/cosmos-sdk/pull/14879) Add `Query/Groups` query to get all the groups.

### Improvements

* (simapp) [#15305](https://github.com/cosmos/cosmos-sdk/pull/15305) Add `AppStateFnWithExtendedCb` with callback function to extend rawState and `AppStateRandomizedFnWithState` with extra genesisState argument which is the genesis state of the app.
* (x/distribution) [#15462](https://github.com/cosmos/cosmos-sdk/pull/15462) Add delegator address to the event for withdrawing delegation rewards
* [#14019](https://github.com/cosmos/cosmos-sdk/issues/14019) Remove the interface casting to allow other implementations of a `CommitMultiStore`.

### Bug Fixes

* (x/auth/vesting) [#15383](https://github.com/cosmos/cosmos-sdk/pull/15383) Add extra checks when creating a periodic vesting account.
* (x/gov) [#13051](https://github.com/cosmos/cosmos-sdk/pull/13051) In SubmitPropsal, when a legacy msg fails it's handler call, wrap the error as ErrInvalidProposalContent (instead of ErrNoProposalHandlerExists).

---

## [v0.46.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.11) - 2023-03-03

### Improvements

* (deps) Migrate to [CometBFT](https://github.com/cometbft/cometbft). Follow the instructions in the [release notes](./RELEASE_NOTES.md).
* (store) [#15152](https://github.com/cosmos/cosmos-sdk/pull/15152) Remove unmaintained and experimental `store/v2alpha1`.
* (store) [#14410](https://github.com/cosmos/cosmos-sdk/pull/14410) `rootmulti.Store.loadVersion` has validation to check if all the module stores' height is correct, it will error if any module store has incorrect height.

### Bug Fixes

* [#15243](https://github.com/cosmos/cosmos-sdk/pull/15243) `LatestBlockResponse` & `BlockByHeightResponse` types' field `sdk_block` was incorrectly cast `proposer_address` bytes to validator operator address, now to consensus address.

