## [v0.46.8-pio-3](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.8-pio-3) - 2023-02-06

### Improvements

* [#498](https://github.com/provenance-io/cosmos-sdk/pull/498) Bump Tendermint to v0.34.25 (from v0.34.24).
* [#499](https://github.com/provenance-io/cosmos-sdk/pull/499) Fix a few listener proto comments.

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.8-pio-2...v0.46.8-pio-3
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.8..v0.46.8-pio-3

---

## [v0.46.8-pio-2](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.8-pio-2) - 2023-02-01

### Features

* [#404](https://github.com/provenance-io/cosmos-sdk/pull/404) Add ADR-038 State Listening with Go Plugin System

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.8-pio-1...v0.46.8-pio-2
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.8..v0.46.8-pio-2

---

## [v0.46.8-pio-1](https://github.com/provenance-io/cosmos-sdk/releases/tag/v0.46.8-pio-1) - 2023-01-25

### Features

* [#270](https://github.com/provenance-io/cosmos-sdk/pull/270) Add functionality to update denom metadata via gov proposal.
* (x/gov,cli) [#434](https://github.com/provenance-io/cosmos-sdk/pull/434) Added `AddGovPropFlagsToCmd` and `ReadGovPropFlags` functions.
* (quarantine) [#335](https://github.com/provenance-io/cosmos-sdk/pull/335) Create the `x/quarantine` module.
* (sanction) [#401](https://github.com/provenance-io/cosmos-sdk/pull/401) Create the `x/sanction` module.

### Improvements

* [#441](https://github.com/provenance-io/cosmos-sdk/pull/441) Bring in Cosmos-SDK [v0.46.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.8) changes.

### Bug Fixes

* [#12184](https://github.com/cosmos/cosmos-sdk/pull/12184) Pull in Cosmos-SDK authz validate basic fix.
* [#444](https://github.com/provenance-io/cosmos-sdk/pull/444) Revert [#13881](https://github.com/cosmos/cosmos-sdk/pull/13881) "Optimize iteration on nested cached KV stores and other operations in general" due to a concurrent iterator issue: [#14786](https://github.com/cosmos/cosmos-sdk/issues/14786).

### Full Commit History

* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.7-pio-1...v0.46.8-pio-1
* https://github.com/provenance-io/cosmos-sdk/compare/v0.46.8..v0.46.8-pio-1

---

## [v0.46.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.8) - 2022-01-23

This release introduces bug fixes and improvements. Notably, the SDK have now switched to Informal Systems' Tendermint fork.
Their fork has no changes compared to the upstream Tendermint, but it is now [maintained by Informal Systems](https://twitter.com/informalinc/status/1613580954383040512). Chains are invited to do the same.

Moreover, this release contains a store fix. The changes have been tested against a v0.46.x chain mainnet with no issues. However, there is a low probability of an edge case happening. Hence, it is recommended to do a **coordinated upgrade** to avoid any issues.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.7...v0.46.8

**NOTE**: The changes mentioned in `v0.46.3` are no longer required. The following replace directive can be removed from the chains.

```go
# Can be deleted from go.mod
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

Instead, `github.com/confio/ics23/go` must be **bumped to `v0.9.0`**.

### Improvements

* [#13881](https://github.com/cosmos/cosmos-sdk/pull/13881) Optimize iteration on nested cached KV stores and other operations in general.
* (x/gov) [#14347](https://github.com/cosmos/cosmos-sdk/pull/14347) Support `v1.Proposal` message in `v1beta1.Proposal.Content`.
* (deps) Use Informal System fork of Tendermint version to [v0.34.24](https://github.com/informalsystems/tendermint/releases/tag/v0.34.24).

### Bug Fixes

* (x/group) [#14526](https://github.com/cosmos/cosmos-sdk/pull/14526) Fix wrong address set in `EventUpdateGroupPolicy`.
* (ante) [#14448](https://github.com/cosmos/cosmos-sdk/pull/14448) Return anteEvents when postHandler fail.

### API Breaking

* (x/gov) [#14422](https://github.com/cosmos/cosmos-sdk/pull/14422) Remove `Migrate_V046_6_To_V046_7` function which shouldn't be used for chains which already migrated to 0.46.
