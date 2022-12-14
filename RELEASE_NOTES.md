## [v0.46.7-pio-1](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.7-pio-1) - 2022-12-14

### Improvements

* Bring in Cosmos-SDK [v0.46.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.7) changes [#402](https://github.com/provenance-io/cosmos-sdk/pull/402).

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.6-pio-1...v0.46.7-pio-1
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.7...v0.46.7-pio-1

---

## [v0.46.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.7) - 2022-12-13

This release introduces bug fixes and improvements. Notably, the upgrade to Tendermint [v0.34.24](https://github.com/tendermint/tendermint/releases/tag/v0.34.24).

Please read the release notes of [v0.46.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.5) if you are upgrading from `<=0.46.4`.

A critical vulnerability has been fixed in the group module. For safety, `v0.46.5` and `v0.46.6` are retracted, even though chains not using the group module are not affected. When using the group module, please upgrade immediately to `v0.46.7`.

An issue has been discovered in the gov module's votes migration. It does not impact proposals and votes tallying, but the gRPC queries on votes are incorrect. This issue is fixed in `v0.46.7`, however:
- if your chain is already on v0.46 using `<= v0.46.6`, a **coordinated upgrade** to v0.46.7 is required. You can use the helper function `govv046.Migrate_V046_6_To_V046_7` for migrating a chain **already on v0.46 with versions <=v0.46.6** to the latest v0.46.7 correct state.
- if your chain is on a previous version <= v0.45, then simply use v0.46.7 when upgrading to v0.46.

**NOTE**: The changes mentioned in `v0.46.3` are no longer required. The following replace directive can be removed from the chains.

```go
# Can be deleted from go.mod
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

Instead, `github.com/confio/ics23/go` must be **bumped to `v0.9.0`**.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.6...v0.46.7

---

### Features

* (client) [#14051](https://github.com/cosmos/cosmos-sdk/pull/14051) Add `--grpc` client option.

### Improvements

* (deps) Bump Tendermint version to [v0.34.24](https://github.com/tendermint/tendermint/releases/tag/v0.34.24).
* [#13651](https://github.com/cosmos/cosmos-sdk/pull/13651) Update `server/config/config.GetConfig` function.
* [#13781](https://github.com/cosmos/cosmos-sdk/pull/13781) Remove `client/keys.KeysCdc`.
* [#14175](https://github.com/cosmos/cosmos-sdk/pull/14175) Add `server.DefaultBaseappOptions(appopts)` function to reduce boiler plate in root.go. 

### State Machine Breaking

* (x/gov) [#14214](https://github.com/cosmos/cosmos-sdk/pull/14214) Fix gov v0.46 migration to v1 votes.
  * Also provide a helper function `govv046.Migrate_V0466_To_V0467` for migrating a chain already on v0.46 with versions <=v0.46.6 to the latest v0.46.7 correct state.
* (x/group) [#14071](https://github.com/cosmos/cosmos-sdk/pull/14071) Don't re-tally proposal after voting period end if they have been marked as ACCEPTED or REJECTED.

### API Breaking Changes

* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Update State Streaming APIs:
    * Add method `ListenCommit` to `ABCIListener`
    * Move `ListeningEnabled` and  `AddListener` methods to `CommitMultiStore`
    * Remove `CacheWrapWithListeners` from `CacheWrap` and `CacheWrapper` interfaces
    * Remove listening APIs from the caching layer (it should only listen to the `rootmulti.Store`)
    * Add three new options to file streaming service constructor.
    * Modify `ABCIListener` such that any error from any method will always halt the app via `panic`
* (store) [#13529](https://github.com/cosmos/cosmos-sdk/pull/13529) Add method `LatestVersion` to `MultiStore` interface, add method `SetQueryMultiStore` to baesapp to support alternative `MultiStore` implementation for query service.

### Bug Fixes

* (baseapp) [#13983](https://github.com/cosmos/cosmos-sdk/pull/13983) Don't emit duplicate ante-handler events when a post-handler is defined.
* (baseapp) [#14049](https://github.com/cosmos/cosmos-sdk/pull/14049) Fix state sync when interval is zero.
* (store) [#13516](https://github.com/cosmos/cosmos-sdk/pull/13516) Fix state listener that was observing writes at wrong time.

